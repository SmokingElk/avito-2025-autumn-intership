package interfaces

import (
	"context"

	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
)

type TeamRepo interface {
	Upsert(ctx context.Context, team teamEntity.Team) error
	GetByName(ctx context.Context, name string) (teamEntity.Team, error)
}
