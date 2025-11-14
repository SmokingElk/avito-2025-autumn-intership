package pullrequestrepopg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
	prErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/pull-request/dto"
	"github.com/jmoiron/sqlx"
)

type PullRequestRepoPg struct {
	db *sqlx.DB
}

func CreatePullRequestRepoPg(db *sqlx.DB) interfaces.PullRequestRepo {
	return &PullRequestRepoPg{
		db: db,
	}
}

func (r *PullRequestRepoPg) GetByReviewer(ctx context.Context, reviewerId string) ([]prEntity.PullRequest, error) {
	query := `
	SELECT
		id,
		pr_name,
		author_id,
		pr_status,
		created_at,
		merged_at,
		reviewers
	FROM pr_with_members WHERE $1 = ANY(reviewers)
	`

	var res []dto.PullRequestDTO

	if err := r.db.SelectContext(ctx, &res, query, reviewerId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []prEntity.PullRequest{}, nil
		}

		return []prEntity.PullRequest{}, fmt.Errorf("failed to select PRs by reviewer: %w", err)
	}

	return []prEntity.PullRequest{}, nil
}

func (r *PullRequestRepoPg) Create(
	ctx context.Context,
	pr prEntity.PullRequest,
	assign interfaces.AssignHandler,
) (prEntity.PullRequest, error) {
	tx, err := r.db.Beginx()

	if err != nil {
		return prEntity.PullRequest{}, fmt.Errorf("failed to begin tx while create pr in postgres: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var team struct {
		id *string `db:"team_id"`
	}

	query := "SELECT team_id FROM member WHERE id = $1"

	if err = tx.GetContext(ctx, &team, query, pr.Id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, prErrors.ErrTeamOrUserNotFound
		}

		return prEntity.PullRequest{}, fmt.Errorf("failed to get team while create pr: %w", err)
	}

	if team.id == nil {
		err = prErrors.ErrTeamOrUserNotFound
		return prEntity.PullRequest{}, err
	}

	query = `
	INSERT INTO pull_request(id, pr_name, author_id, team_id, pr_status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err = tx.ExecContext(
		ctx,
		query,
		pr.Id,
		pr.Name,
		pr.AuthorId,
		team.id,
		string(pr.Status),
		pr.CreatedAt,
	); err != nil {
		return prEntity.PullRequest{}, fmt.Errorf("failed to create pr in postgres: %w", err)
	}

	var members []struct {
		id string `db:"id"`
	}

	query = "SELECT id FROM member WHERE team_id = $1"

	if err = tx.SelectContext(ctx, &members, query, team.id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, fmt.Errorf("failed to get team members while create pr")
		}
	}

	membersIds := make([]string, 0, len(members))

	for _, member := range members {
		membersIds = append(membersIds, member.id)
	}

	assigned := assign(pr.AuthorId, membersIds)

	for _, reviewer := range assigned {
		query = `
		INSERT INTO assigned_reviewer(member_id, pr_id) 
		VALUES ($1, $2)
		`

		if _, err = tx.ExecContext(ctx, query, reviewer, pr.Id); err != nil {
			return prEntity.PullRequest{}, fmt.Errorf("failed to add pr reviewer to postgres: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return prEntity.PullRequest{}, fmt.Errorf("failed to commit tx while create pr postgres: %w", err)
	}

	pr.Reviewers = assigned

	return pr, nil
}

func (r *PullRequestRepoPg) Merge(ctx context.Context, prId string) (time.Time, error) {
	tx, err := r.db.Beginx()

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to begin tx while merge pr in postgres: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
	SELECT
		id,
		pr_status,
		merged
	FROM pull_request WHERE id = $1
	`

	var pr dto.PullRequestDTO

	if err = tx.GetContext(ctx, &pr, query, prId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, prErrors.ErrNotFound
		}

		return time.Time{}, fmt.Errorf("failed to get pr while merge: %w", err)
	}

	mergeTime := time.Now()

	if pr.Status == "MERGED" {
		mergeTime = pr.MergedAt
	} else {
		query = `
		UPDATE pull_request
		SET status = 'MERGED', merged_at = $1 
		WHERE id = $2
		`

		if _, err = tx.ExecContext(ctx, query, mergeTime, prId); err != nil {
			return time.Time{}, fmt.Errorf("failed to update status while merge mr: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return time.Time{}, fmt.Errorf("failed to commit tx while merge pr postgres: %w", err)
	}

	return mergeTime, nil
}

func (r *PullRequestRepoPg) Reassign(
	ctx context.Context,
	prId string,
	oldReviewerId string,
	assign interfaces.ReassignHandler,
) (prEntity.PullRequest, string, error) {
	tx, err := r.db.Beginx()

	if err != nil {
		return prEntity.PullRequest{}, "", fmt.Errorf("failed to begin tx while create pr in postgres: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var pr dto.PullRequestDTO

	query := `
	SELECT
		pr_name,
		author_id,
		pr_status,
		created_at,
		merged_at,
		reviewers
	FROM pr_with_members WHERE id = $1
	`

	pr.Id = prId

	if err = tx.GetContext(ctx, &pr, query, prId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, "", prErrors.ErrNotFound
		}

		return prEntity.PullRequest{}, "", fmt.Errorf("failed to get pr to reassign: %w", err)
	}

	var teamMembers []struct {
		id string `db:"member_id"`
	}

	query = "SELECT member_id FROM assigned_reviewer WHERE pr_id = $1"

	if err = tx.SelectContext(ctx, &teamMembers, query, prId); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, "", fmt.Errorf("failed to get current reviewers while reassign: %w", err)
		}
	}

	teamMembersIds := make([]string, 0, len(teamMembers))

	for _, member := range teamMembers {
		teamMembersIds = append(teamMembersIds, member.id)
	}

	newReviewer, err := assign(pr.AuthorId, pr.Reviewers, teamMembersIds)

	if err != nil {
		return prEntity.PullRequest{}, "", err
	}

	query = "DELETE FROM assigned_reviewer WHERE member_id = $1"

	if _, err = tx.ExecContext(ctx, query, oldReviewerId); err != nil {
		return prEntity.PullRequest{}, "", fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	query = "INSERT INTO assigned_reviewer(member_id, pr_id) VALUES ($1, $2)"

	if _, err = tx.ExecContext(ctx, query, newReviewer, prId); err != nil {
		return prEntity.PullRequest{}, "", fmt.Errorf("failed to remove old reviewer")
	}

	if err = tx.Commit(); err != nil {
		return prEntity.PullRequest{}, "", fmt.Errorf("failed to commit tx while merge pr postgres: %w", err)
	}

	for i, reviewer := range pr.Reviewers {
		if reviewer == oldReviewerId {
			pr.Reviewers[i] = newReviewer
		}
	}

	return prEntity.PullRequest{}, newReviewer, nil
}
