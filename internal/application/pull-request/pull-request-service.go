package pullrequestservice

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
	prErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
)

type PullRequestService struct {
	repo interfaces.PullRequestRepo
	cfg  *config.PullRequestConfig
}

func CreatePullRequestService(repo interfaces.PullRequestRepo, cfg *config.PullRequestConfig) interfaces.PullRequestService {
	return &PullRequestService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *PullRequestService) GetByReviewer(ctx context.Context, reviewerId string) ([]prEntity.PullRequest, error) {
	prs, err := s.repo.GetByReviewer(ctx, reviewerId, s.cfg.OutLimit)

	if err != nil {
		return []prEntity.PullRequest{}, fmt.Errorf("failed to get pull requests from repo: %w", err)
	}

	return prs, nil
}

func (s *PullRequestService) Create(ctx context.Context, prId, prName, authorId string) (prEntity.PullRequest, error) {
	pr := prEntity.NewPullRequest(prId, prName, authorId)

	prWithReviewers, err := s.repo.Create(ctx, pr, func(authorId string, members []memberEntity.Member) []string {
		activeMembers := make([]string, 0, len(members))

		for _, member := range members {
			if member.Activity == memberEntity.MemberActive && member.Id != authorId {
				activeMembers = append(activeMembers, member.Id)
			}
		}

		rand.Shuffle(len(activeMembers), func(i, j int) {
			activeMembers[i], activeMembers[j] = activeMembers[j], activeMembers[i]
		})

		return activeMembers[:min(s.cfg.TargetReviewersCount, len(activeMembers))]
	})

	if err != nil {
		if errors.Is(err, prErrors.ErrAlreadyExists) {
			return prEntity.PullRequest{}, err
		}

		return prEntity.PullRequest{}, fmt.Errorf("failed to store pr in repo: %w", err)
	}

	return prWithReviewers, nil
}

func (s *PullRequestService) Merge(ctx context.Context, prId string) (prEntity.PullRequest, error) {
	mergedPr, err := s.repo.UpdateStatus(ctx, prId, func(pr prEntity.PullRequest) (prEntity.PullRequest, bool) {
		if pr.Status == prEntity.PRMerged {
			return pr, false
		}

		pr.MergedAt = time.Now()
		pr.Status = prEntity.PRMerged
		return pr, true
	})

	if err != nil {
		if errors.Is(err, prErrors.ErrNotFound) {
			return prEntity.PullRequest{}, err
		}

		return prEntity.PullRequest{}, fmt.Errorf("failed to merge pr in repo: %w", err)
	}

	return mergedPr, nil
}

func (s *PullRequestService) Reassign(
	ctx context.Context,
	prId string,
	oldReviewerId string,
) (prEntity.PullRequest, string, error) {
	updatedPr, newReviewer, err := s.repo.Reassign(
		ctx,
		prId,
		oldReviewerId,
		func(authorId string, pr prEntity.PullRequest, teamMembers []memberEntity.Member) (string, error) {
			if pr.Status == prEntity.PRMerged {
				return "", prErrors.ErrAlreadyMerged
			}

			currentReviewersMap := make(map[string]struct{})

			for _, member := range pr.Reviewers {
				currentReviewersMap[member] = struct{}{}
			}

			if _, ok := currentReviewersMap[oldReviewerId]; !ok {
				return "", prErrors.ErrTeamOrUserNotFound
			}

			activeMembers := make([]string, 0, len(teamMembers))

			for _, member := range teamMembers {
				if member.Activity != memberEntity.MemberActive || member.Id == authorId {
					continue
				}

				if _, ok := currentReviewersMap[member.Id]; ok {
					continue
				}

				activeMembers = append(activeMembers, member.Id)
			}

			if len(activeMembers) == 0 {
				return "", prErrors.ErrCannotReassign
			}

			idx := rand.Intn(len(activeMembers))

			return activeMembers[idx], nil
		},
	)

	if err != nil {
		if errors.Is(err, prErrors.ErrCannotReassign) ||
			errors.Is(err, prErrors.ErrTeamOrUserNotFound) ||
			errors.Is(err, prErrors.ErrNotFound) ||
			errors.Is(err, prErrors.ErrAlreadyMerged) {

			return prEntity.PullRequest{}, "", err
		}

		return prEntity.PullRequest{}, "", fmt.Errorf("failed to reassign reviewer in repo: %w", err)
	}

	return updatedPr, newReviewer, nil
}
