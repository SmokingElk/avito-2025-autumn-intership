package interfaces

import (
	"context"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/entity"
)

type StatsRepo interface {
	GetAssigmentsPerMember(ctx context.Context, limit, offset int) ([]entity.AssignmentsPerMember, error)
}
