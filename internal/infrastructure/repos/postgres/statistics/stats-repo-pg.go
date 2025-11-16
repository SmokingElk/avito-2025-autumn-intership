package statsrepopg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/entity"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/postgres/statistics/dto"
	"github.com/jmoiron/sqlx"
)

type StatsRepoPg struct {
	db *sqlx.DB
}

func CreateStatsRepoPg(db *sqlx.DB) interfaces.StatsRepo {
	return &StatsRepoPg{
		db: db,
	}
}

func (r *StatsRepoPg) GetAssigmentsPerMember(ctx context.Context, limit, offset int) ([]entity.AssignmentsPerMember, error) {
	query := `
	SELECT member_id, COUNT(*) AS assigments_count
	FROM assigned_reviewer
	GROUP BY member_id
	ORDER BY assigments_count
	LIMIT $1
	OFFSET $2
	`

	var stats []dto.AssignmentsPerMember

	if err := r.db.SelectContext(ctx, &stats, query, limit, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []entity.AssignmentsPerMember{}, nil
		}

		return []entity.AssignmentsPerMember{}, fmt.Errorf("failed to get assigments from postgres: %w", err)
	}

	res := make([]entity.AssignmentsPerMember, 0, len(stats))

	for _, assignmentsPerMember := range stats {
		res = append(res, assignmentsPerMember.ToAssignmentsPerMemberEntity())
	}

	return res, nil
}
