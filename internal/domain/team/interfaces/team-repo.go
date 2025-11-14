package interfaces

import (
	"context"

	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
)

// extract matching logic from infrastructure layer
type TeamMatcher func(currentTeam teamEntity.Team) bool

type TeamRepo interface {
	Upsert(ctx context.Context, team teamEntity.Team, matcher TeamMatcher) error
	GetByName(ctx context.Context, name string) (teamEntity.Team, error)
}
