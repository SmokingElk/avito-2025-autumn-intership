package statsservice

import (
	"context"
	"fmt"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/entity"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/interfaces"
)

type StatsService struct {
	repo interfaces.StatsRepo
}

func CreateStatsService(repo interfaces.StatsRepo) interfaces.StatsService {
	return &StatsService{
		repo: repo,
	}
}

func (s *StatsService) GetAssignmentsPerMember(ctx context.Context, limit, offset int) ([]entity.AssignmentsPerMember, error) {
	stats, err := s.repo.GetAssignmentsPerMember(ctx, limit, offset)

	if err != nil {
		return []entity.AssignmentsPerMember{}, fmt.Errorf("failed to get assignments stats from repo: %w", err)
	}

	return stats, nil
}
