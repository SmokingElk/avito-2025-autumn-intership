package teamrepopg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
	teamErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/team/dto"
	"github.com/jmoiron/sqlx"
)

type TeamRepoPg struct {
	db *sqlx.DB
}

func CreateTeamRepoPg(db *sqlx.DB) interfaces.TeamRepo {
	return &TeamRepoPg{
		db: db,
	}
}

func (r *TeamRepoPg) Upsert(ctx context.Context, team teamEntity.Team, matcher interfaces.TeamMatcher) error {
	tx, err := r.db.Beginx()

	if err != nil {
		return fmt.Errorf("failed to begin tx while upsert team to postgres: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	updateTeam := true
	currentTeam, err := r.getTeamWithMembers(ctx, tx, team.Name)

	if err != nil {
		if err != teamErrors.ErrTeamNotFound {
			return err
		}

		updateTeam = false
	} else {
		team.Id = currentTeam.Id
	}

	if matcher(currentTeam) {
		err = teamErrors.ErrTeamExists
		return err
	}

	query := `
	INSERT INTO team(id, name) VALUES ($1, $2) 
	ON CONFLICT (id, name) 
	DO NOTHING
	`

	if _, err = tx.ExecContext(ctx, query, team.Id, team.Name); err != nil {
		return fmt.Errorf("failed to upsert team into postgres table: %w", err)
	}

	newMembers := make(map[string]struct{})

	// upsert new members if they are not members of other team
	for _, member := range team.Members {
		query = `SELECT id FROM member WHERE id = $1 AND team_id IS NOT NULL`

		if _, err = tx.ExecContext(ctx, query, member.Id); err == nil {
			err = teamErrors.ErrMemberOfOtherTeam
			return err
		}

		query := `
		INSERT INTO member(id, username, activity, team_id) VALUES ($1, $2, $3, $4) 
		ON CONFLICT DO 
		UPDATE member SET team_id = $4 WHERE id = $1
		`

		if _, err = tx.ExecContext(ctx, query, member.Id, member.Username, string(member.Activity), team.Id); err != nil {
			return fmt.Errorf("failed to upsert member of team into postgres table: %w", err)
		}

		newMembers[member.Id] = struct{}{}
	}

	if updateTeam {
		// delete old members
		for _, oldMember := range currentTeam.Members {
			if _, ok := newMembers[oldMember.Id]; ok {
				continue
			}

			// delete member from reviewers of opened PR
			query = `
			DELETE a 
			FROM assigned_reviewers AS a
			INNER JOIN pull_request AS pr 
				ON a.pr_id = pr.id
			WHERE pr.status = 'OPEN' AND a.member_id = $1
			`

			if _, err = tx.ExecContext(ctx, query, oldMember.Id); err != nil {
				return fmt.Errorf("failed to remove member from reviewers: %w", err)
			}

			query = "UPDATE member SET team_id = NULL WHERE id = $1"

			if _, err = tx.ExecContext(ctx, query, oldMember.Id); err != nil {
				return fmt.Errorf("failed to remove member from team members: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx while upsert team postgres: %w", err)
	}

	return nil
}

func (r *TeamRepoPg) GetByName(ctx context.Context, name string) (teamEntity.Team, error) {
	tx, err := r.db.Beginx()

	if err != nil {
		return teamEntity.Team{}, fmt.Errorf("failed to begin tx while getting team from postgres: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	res, err := r.getTeamWithMembers(ctx, tx, name)

	if err != nil {
		return teamEntity.Team{}, err
	}

	if err = tx.Commit(); err != nil {
		return teamEntity.Team{}, fmt.Errorf("failed to commit tx while getting team from postgres: %w", err)
	}

	return res, nil
}

func (r *TeamRepoPg) getTeamWithMembers(ctx context.Context, tx *sqlx.Tx, name string) (teamEntity.Team, error) {
	query := "SELECT id, name FROM team WHERE name = $1"

	var team dto.TeamDTO
	if err := tx.GetContext(ctx, &team, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return teamEntity.Team{}, teamErrors.ErrTeamNotFound
		}

		return teamEntity.Team{}, fmt.Errorf("failed to select team from postgres table: %w", err)
	}

	query = "SELECT id, name, activity, team_id FROM member WHERE team_id = $1"

	if err := tx.SelectContext(ctx, &team.Members, query, team.Id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return teamEntity.Team{}, fmt.Errorf("failed to select members of team from postgres table: %w", err)
		}
	}

	return team.ToTeamEntity(), nil
}
