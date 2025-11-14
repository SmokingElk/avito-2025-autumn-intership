package interfaces

import (
	"context"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
)

type MemberRepo interface {
	SetActivity(ctx context.Context, userId string, activity memberEntity.MemberActivity) (memberEntity.Member, error)
}
