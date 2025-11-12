package interfaces

import (
	"context"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
)

type TeamService interface {
	Add(ctx context.Context, name string, membersList []memberEntity.Member) error
	GetByName(ctx context.Context, name string) (teamEntity.Team, error)
}
