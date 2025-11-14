package interfaces

import (
	"context"

	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
)

// extract assign logic from infrastructure layer
type AssignHandler func(authorId string, members []memberEntity.Member) []string
type ReassignHandler func(authorId string, currentReviewers []string, teamMembers []memberEntity.Member) (string, error)

type PullRequestRepo interface {
	GetByReviewer(ctx context.Context, reviewerId string, limit int) ([]prEntity.PullRequest, error)
	Create(ctx context.Context, pr prEntity.PullRequest, assign AssignHandler) (prEntity.PullRequest, error)
	Merge(ctx context.Context, prId string) (prEntity.PullRequest, error)
	Reassign(
		ctx context.Context,
		prId string,
		oldReviewerId string,
		assign ReassignHandler,
	) (prEntity.PullRequest, string, error)
}
