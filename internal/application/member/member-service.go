package memberservice

import (
	"context"
	"errors"
	"fmt"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	memberErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/interfaces"
)

type MemberService struct {
	repo interfaces.MemberRepo
}

func CreateMemberService(repo interfaces.MemberRepo) interfaces.MemberService {
	return &MemberService{
		repo: repo,
	}
}

func (s *MemberService) SetIsActive(ctx context.Context, userId string, isActive bool) error {
	activity := memberEntity.MemberInactive
	if isActive {
		activity = memberEntity.MemberActive
	}

	err := s.repo.SetActivity(ctx, userId, activity)

	if err != nil {
		if errors.Is(err, memberErrors.ErrMemberNotFound) {
			return err
		}

		return fmt.Errorf("failed to set active in repo: %w", err)
	}

	return nil
}
