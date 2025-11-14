package pullrequestrepopg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
	prErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/pull-request/dto"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const primaryKeyViolation = "23505"

type PullRequestRepoPg struct {
	db *sqlx.DB
}

func CreatePullRequestRepoPg(db *sqlx.DB) interfaces.PullRequestRepo {
	return &PullRequestRepoPg{
		db: db,
	}
}

func (r *PullRequestRepoPg) GetByReviewer(ctx context.Context, reviewerId string, limit int) ([]prEntity.PullRequest, error) {
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
	ORDER BY created_at
	LIMIT $2
	`

	var prs []dto.PullRequestDTO

	if err := r.db.SelectContext(ctx, &prs, query, reviewerId, limit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []prEntity.PullRequest{}, nil
		}

		return []prEntity.PullRequest{}, fmt.Errorf("failed to select PRs by reviewer: %w", err)
	}

	res := make([]prEntity.PullRequest, 0, len(prs))

	for _, pr := range prs {
		res = append(res, pr.ToPullRequestEntity())
	}

	return res, nil
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
		Id *string `db:"team_id"`
	}

	query := "SELECT team_id FROM team_member WHERE id = $1"

	if err = tx.GetContext(ctx, &team, query, pr.AuthorId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, prErrors.ErrTeamOrUserNotFound
		}

		return prEntity.PullRequest{}, fmt.Errorf("failed to get team while create pr: %w", err)
	}

	if team.Id == nil {
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
		team.Id,
		string(pr.Status),
		pr.CreatedAt,
	); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == primaryKeyViolation {
				return prEntity.PullRequest{}, prErrors.ErrAlreadyExists
			}
		}

		return prEntity.PullRequest{}, fmt.Errorf("failed to create pr in postgres: %w", err)
	}

	var members []dto.MemberDTO

	query = "SELECT id, activity FROM team_member WHERE team_id = $1"

	if err = tx.SelectContext(ctx, &members, query, team.Id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, fmt.Errorf("failed to get team members while create pr")
		}
	}

	membersEntities := make([]memberEntity.Member, 0, len(members))

	for _, member := range members {
		membersEntities = append(membersEntities, member.ToMemberEntity())
	}

	assigned := assign(pr.AuthorId, membersEntities)

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

func (r *PullRequestRepoPg) UpdateStatus(
	ctx context.Context,
	prId string,
	updateStatusHandler interfaces.UpdateStatusHandler,
) (prEntity.PullRequest, error) {
	tx, err := r.db.Beginx()

	if err != nil {
		return prEntity.PullRequest{}, fmt.Errorf("failed to begin tx while merge pr in postgres: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
	SELECT
		id,
		pr_name,
		author_id,
		pr_status,
		created_at,
		merged_at,
		reviewers
	FROM pr_with_members WHERE id = $1
	`

	var pr dto.PullRequestDTO

	if err = tx.GetContext(ctx, &pr, query, prId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, prErrors.ErrNotFound
		}

		return prEntity.PullRequest{}, fmt.Errorf("failed to get pr while merge: %w", err)
	}

	pr.Id = prId
	prUpdated, updated := updateStatusHandler(pr.ToPullRequestEntity())

	if updated {
		query = `
		UPDATE pull_request
		SET pr_status = $1, merged_at = $2 
		WHERE id = $3
		`

		if _, err = tx.ExecContext(ctx, query, string(prUpdated.Status), prUpdated.MergedAt, prId); err != nil {
			return prEntity.PullRequest{}, fmt.Errorf("failed to update status while merge mr: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return prEntity.PullRequest{}, fmt.Errorf("failed to commit tx while merge pr postgres: %w", err)
	}

	return prUpdated, nil
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
		team_id,
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

	var teamMembers []dto.MemberDTO

	query = `
	SELECT id, activity 
	FROM team_member 
	WHERE team_id = $1
	`

	if err = tx.SelectContext(ctx, &teamMembers, query, pr.TeamId); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return prEntity.PullRequest{}, "", fmt.Errorf("failed to get current reviewers while reassign: %w", err)
		}
	}

	teamMembersEntities := make([]memberEntity.Member, 0, len(teamMembers))

	for _, member := range teamMembers {
		teamMembersEntities = append(teamMembersEntities, member.ToMemberEntity())
	}

	newReviewer, err := assign(pr.AuthorId, pr.ToPullRequestEntity(), teamMembersEntities)

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

	return pr.ToPullRequestEntity(), newReviewer, nil
}
