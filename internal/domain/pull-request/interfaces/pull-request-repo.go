package interfaces

import (
	"context"

	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
)

// extract assign logic from infrastructure layer
type AssignHandler func(members []string) []string

type PullRequestRepo interface {
	GetByReviewer(ctx context.Context, reviewerId string) ([]prEntity.PullRequest, error)
	Create(ctx context.Context, pr prEntity.PullRequest, assign AssignHandler) (prEntity.PullRequest, error)
	Merge(ctx context.Context, prId string) error
	Reassign(ctx context.Context, prId string, assign AssignHandler) (prEntity.PullRequest, error)
}
