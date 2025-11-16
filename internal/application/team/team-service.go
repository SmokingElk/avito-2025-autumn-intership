package teamservice

import (
	"context"
	"errors"
	"fmt"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
	teamErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/interfaces"
)

type TeamService struct {
	repo interfaces.TeamRepo
}

func CreateTeamService(repo interfaces.TeamRepo) interfaces.TeamService {
	return &TeamService{
		repo: repo,
	}
}

func (s *TeamService) Upsert(ctx context.Context, name string, membersList []memberEntity.Member) error {
	team := teamEntity.NewTeam(name, membersList)

	newMembers := make(map[string]struct{})
	for _, member := range membersList {
		newMembers[member.Id] = struct{}{}
	}

	err := s.repo.Upsert(ctx, team, func(currentTeam teamEntity.Team) bool {
		if len(newMembers) != len(currentTeam.Members) {
			return false
		}

		for _, member := range currentTeam.Members {
			if _, ok := newMembers[member.Id]; !ok {
				return false
			}
		}

		return true
	})

	if err != nil {
		if errors.Is(err, teamErrors.ErrTeamExists) || errors.Is(err, teamErrors.ErrMemberOfOtherTeam) {
			return err
		}

		return fmt.Errorf("failed to upsert team to repo: %w", err)
	}

	return nil
}

func (s *TeamService) GetByName(ctx context.Context, name string) (teamEntity.Team, error) {
	team, err := s.repo.GetByName(ctx, name)

	if err != nil {
		if errors.Is(err, teamErrors.ErrTeamNotFound) {
			return teamEntity.Team{}, err
		}

		return teamEntity.Team{}, fmt.Errorf("failed to get team from repo: %w", err)
	}

	return team, nil
}

func (s *TeamService) DeactivateAll(ctx context.Context, name string) error {
	if err := s.repo.SetActivityForAll(ctx, name, memberEntity.MemberInactive); err != nil {
		if errors.Is(err, teamErrors.ErrTeamNotFound) {
			return err
		}

		return fmt.Errorf("failed to deactivate in repo: %w", err)
	}

	return nil
}
