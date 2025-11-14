package memberrepopg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	memberErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/interfaces"
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

func (r *MemberRepoPg) SetActivity(ctx context.Context, userId string, activity memberEntity.MemberActivity) error {
	query := "UPDATE member SET activity = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, string(activity), userId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return memberErrors.ErrMemberNotFound
		}

		return fmt.Errorf("failed to update member activity in postgres: %w", err)
	}

	return nil
}
