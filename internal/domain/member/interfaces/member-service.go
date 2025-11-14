package interfaces

import (
	"context"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
)

type MemberService interface {
	SetIsActive(ctx context.Context, userId string, isActive bool) (entity.Member, error)
}
