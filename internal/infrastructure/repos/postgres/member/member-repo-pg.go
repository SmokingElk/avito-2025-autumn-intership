package memberrepopg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	memberErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/postgres/member/dto"
	"github.com/jmoiron/sqlx"
)

type MemberRepoPg struct {
	db *sqlx.DB
}

func CreateMemberRepoPg(db *sqlx.DB) interfaces.MemberRepo {
	return &MemberRepoPg{
		db: db,
	}
}

func (r *MemberRepoPg) SetActivity(
	ctx context.Context,
	userId string,
	activity memberEntity.MemberActivity,
) (memberEntity.Member, error) {
	tx, err := r.db.Beginx()

	if err != nil {
		return memberEntity.Member{}, fmt.Errorf("failed to begin tx while set activity to postgres: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var member dto.MemberDTO

	query := `
	SELECT m.id, m.username, m.activity, t.team_name
	FROM team_member AS m
	LEFT JOIN team AS t
		ON m.team_id = t.id
	WHERE m.id = $1
	`

	if err = tx.GetContext(ctx, &member, query, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return memberEntity.Member{}, memberErrors.ErrMemberNotFound
		}

		return memberEntity.Member{}, fmt.Errorf("failed to get member in postgres: %w", err)
	}

	query = "UPDATE team_member SET activity = $1 WHERE id = $2"
	_, err = tx.ExecContext(ctx, query, string(activity), userId)

	if err != nil {
		return memberEntity.Member{}, fmt.Errorf("failed to update member activity in postgres: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return memberEntity.Member{}, fmt.Errorf("failed to commit tx while set activity postgres: %w", err)
	}

	res := member.ToMemberEntity()

	res.Activity = activity

	return res, nil
}
