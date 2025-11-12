package interfaces

import (
	"context"

	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
)

type PullRequestService interface {
	GetByReviewer(ctx context.Context, reviewerId string) error
	Create(ctx context.Context, prId, prName, authorId string) (prEntity.PullRequest, error)
	Merge(ctx context.Context, prId string) error
	Reassign(ctx context.Context, prId string, oldReviewerId string) (prEntity.PullRequest, string, error)
}
