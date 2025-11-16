package teamrepopg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
	teamErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/postgres/team/dto"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type TeamRepoPg struct {
	db     *sqlx.DB
	logger zerolog.Logger
}

func CreateTeamRepoPg(db *sqlx.DB, log zerolog.Logger) interfaces.TeamRepo {
	return &TeamRepoPg{
		db:     db,
		logger: log,
	}
}

func (r *TeamRepoPg) Upsert(ctx context.Context, team teamEntity.Team, matcher interfaces.TeamMatcher) error {
	tx, err := r.db.Beginx()

	if err != nil {
		return fmt.Errorf("failed to begin tx while upsert team to postgres: %w", err)
	}

	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				r.logger.Error().Err(err).Msg("failed to rollback transaction")
			}
		}
	}()

	updateTeam := true
	currentTeam, err := r.getTeamWithMembers(ctx, tx, team.Name)

	if err != nil {
		if !errors.Is(err, teamErrors.ErrTeamNotFound) {
			return err
		}

		updateTeam = false
	} else {
		team.Id = currentTeam.Id
	}

	err = nil

	if updateTeam && matcher(currentTeam) {
		err = teamErrors.ErrTeamExists
		return err
	}

	query := `
	INSERT INTO team(id, team_name) VALUES ($1, $2) 
	ON CONFLICT
	DO NOTHING
	`

	if _, err = tx.ExecContext(ctx, query, team.Id, team.Name); err != nil {
		return fmt.Errorf("failed to upsert team into postgres table: %w", err)
	}

	newMembers := make(map[string]struct{})

	// upsert new members if they are not members of other team
	for _, member := range team.Members {
		var m dto.MemberDTO

		query = `
		SELECT m.id 
		FROM team_member as m
		INNER JOIN team as t 
			ON t.id = m.team_id 
		WHERE m.id = $1 AND t.team_name <> $2
		`

		if err = tx.GetContext(ctx, &m, query, member.Id, team.Name); err == nil {
			err = teamErrors.ErrMemberOfOtherTeam
			return err
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check if member in other team: %w", err)
		}

		query := `
		INSERT INTO team_member(id, username, activity, team_id) VALUES ($1, $2, $3, $4) 
		ON CONFLICT(id) DO UPDATE 
		SET team_id = EXCLUDED.team_id
		WHERE team_member.id = EXCLUDED.id
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
			DELETE FROM assigned_reviewer 
			USING pull_request AS pr
			WHERE assigned_reviewer.pr_id = pr.id 
				AND pr.pr_status = 'OPEN' 
				AND assigned_reviewer.member_id = $1
			`

			if _, err = tx.ExecContext(ctx, query, oldMember.Id); err != nil {
				return fmt.Errorf("failed to remove member from reviewers: %w", err)
			}

			query = "UPDATE team_member SET team_id = NULL WHERE id = $1"

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
			if err := tx.Rollback(); err != nil {
				r.logger.Error().Err(err).Msg("failed to rollback transaction")
			}
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

func (r *TeamRepoPg) SetActivityForAll(ctx context.Context, name string, activity memberEntity.MemberActivity) error {
	query := `
	UPDATE team_member 
	SET activity = $1
	WHERE team_id = (
		SELECT id FROM team WHERE team_name = $2
	)
	RETURNING team_id
	`

	var resp struct {
		Id string `db:"team_id"`
	}

	if err := r.db.GetContext(ctx, &resp, query, string(activity), name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return teamErrors.ErrTeamNotFound
		}

		return fmt.Errorf("failed to set activity for all members of team in postgres: %w", err)
	}

	return nil
}

func (r *TeamRepoPg) getTeamWithMembers(ctx context.Context, tx *sqlx.Tx, name string) (teamEntity.Team, error) {
	query := "SELECT id, team_name FROM team WHERE team_name = $1"

	var team dto.TeamDTO
	if err := tx.GetContext(ctx, &team, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return teamEntity.Team{}, teamErrors.ErrTeamNotFound
		}

		return teamEntity.Team{}, fmt.Errorf("failed to select team from postgres table: %w", err)
	}

	query = "SELECT id, username, activity, team_id FROM team_member WHERE team_id = $1"

	if err := tx.SelectContext(ctx, &team.Members, query, team.Id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return teamEntity.Team{}, fmt.Errorf("failed to select members of team from postgres table: %w", err)
		}
	}

	return team.ToTeamEntity(), nil
}
