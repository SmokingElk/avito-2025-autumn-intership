package interfaces

import "context"

type MemberService interface {
	SetIsActive(ctx context.Context, userId string, isActive bool) error
}
