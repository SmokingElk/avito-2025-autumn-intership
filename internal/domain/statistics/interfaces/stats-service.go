package interfaces

import (
	"context"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/entity"
)

type StatsService interface {
	GetAssignmentsPerMember(ctx context.Context, limit, offset int) ([]entity.AssignmentsPerMember, error)
}
