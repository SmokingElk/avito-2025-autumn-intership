package interfaces

import (
	"context"
	"time"

	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
)

// extract assign logic from infrastructure layer
type AssignHandler func(authorId string, members []string) []string
type ReassignHandler func(authorId string, currentReviewers []string, teamMembers []string) (string, error)

type PullRequestRepo interface {
	GetByReviewer(ctx context.Context, reviewerId string) ([]prEntity.PullRequest, error)
	Create(ctx context.Context, pr prEntity.PullRequest, assign AssignHandler) (prEntity.PullRequest, error)
	Merge(ctx context.Context, prId string) (time.Time, error)
	Reassign(
		ctx context.Context,
		prId string,
		oldReviewerId string,
		assign ReassignHandler,
	) (prEntity.PullRequest, string, error)
}
