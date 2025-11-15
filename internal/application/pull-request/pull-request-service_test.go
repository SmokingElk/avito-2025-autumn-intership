package pullrequestservice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	pullrequestservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/pull-request"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
	prErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
	prMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetByReviewer(t *testing.T) {
	reviewerId := "reviewer"

	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		repoError     error
		expectedPRs   []prEntity.PullRequest
		expectedError string
		noError       bool
	}

	testCases := []testCase{
		{
			what:          "failed to get pull requests from repo",
			repoError:     errors.New("db is down"),
			expectedError: "failed to get pull requests from repo: db is down",
		},

		{
			what:      "successfully get PRs",
			repoError: nil,
			noError:   true,
			expectedPRs: []prEntity.PullRequest{
				{
					Id:       "pr1",
					Name:     "pull request 1",
					AuthorId: "u1",
					Status:   prEntity.PROpen,
					Reviewers: []string{
						"u2",
						"u3",
					},
				},
				{
					Id:       "pr2",
					Name:     "pull request 2",
					AuthorId: "u2",
					Status:   prEntity.PRMerged,
					Reviewers: []string{
						"u4",
						"u3",
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.EXPECT().GetByReviewer(
				gomock.Any(),
				reviewerId,
				config.OutLimit,
			).Return(tc.expectedPRs, tc.repoError)

			service := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			prs, err := service.GetByReviewer(context.Background(), reviewerId)

			if tc.noError {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPRs, prs)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestCreate(t *testing.T) {
	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		prId                    string
		prName                  string
		authorId                string
		teamMembers             []memberEntity.Member
		expectedPR              prEntity.PullRequest
		expectedPRWithReviewers prEntity.PullRequest
		repoError               error
		expectedError           string
		noError                 bool
	}

	testCases := []testCase{
		{
			what: "pr already exists",

			prId:     "pr1",
			prName:   "pull request 1",
			authorId: "u1",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
			},
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			expectedPRWithReviewers: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{},
			},
			repoError:     prErrors.ErrAlreadyExists,
			expectedError: prErrors.ErrAlreadyExists.Error(),
		},

		{
			what: "failed to store pr in repo",

			prId:     "pr1",
			prName:   "pull request 1",
			authorId: "u1",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
			},
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			expectedPRWithReviewers: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{},
			},
			repoError:     errors.New("db is down"),
			expectedError: "failed to store pr in repo: db is down",
		},

		{
			what: "not enough for target count",

			prId:     "pr1",
			prName:   "pull request 1",
			authorId: "u1",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},

				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
			},
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			expectedPRWithReviewers: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2"},
			},
			repoError: nil,
			noError:   true,
		},

		{
			what: "enough for target count",

			prId:     "pr1",
			prName:   "pull request 1",
			authorId: "u1",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},

				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},

				{
					Id:       "u3",
					Activity: memberEntity.MemberActive,
				},
			},
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			expectedPRWithReviewers: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2", "u3"},
			},
			repoError: nil,
			noError:   true,
		},

		{
			what: "ignore inactive",

			prId:     "pr1",
			prName:   "pull request 1",
			authorId: "u1",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},

				{
					Id:       "u2",
					Activity: memberEntity.MemberInactive,
				},

				{
					Id:       "u3",
					Activity: memberEntity.MemberInactive,
				},

				{
					Id:       "u4",
					Activity: memberEntity.MemberActive,
				},
			},
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			expectedPRWithReviewers: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u4"},
			},
			repoError: nil,
			noError:   true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.
				EXPECT().
				Create(gomock.Any(), prEntity.Matcher(tc.expectedPR), gomock.Any()).
				DoAndReturn(func(
					ctx context.Context,
					pr prEntity.PullRequest,
					callback interfaces.AssignHandler,
				) (prEntity.PullRequest, error) {
					reviewers := callback(tc.authorId, tc.teamMembers)
					assert.ElementsMatch(t, tc.expectedPRWithReviewers.Reviewers, reviewers)

					return tc.expectedPRWithReviewers, tc.repoError
				})

			service := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			pr, err := service.Create(context.Background(), tc.prId, tc.prName, tc.authorId)

			if tc.noError {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPRWithReviewers, pr)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestMerge(t *testing.T) {
	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		prId            string
		storedPr        prEntity.PullRequest
		expectedPr      prEntity.PullRequest
		expectedUpdated bool
		repoError       error
		expectedError   string
		noError         bool
	}

	testCases := []testCase{
		{
			what: "pr not found",

			prId: "pr1",
			storedPr: prEntity.PullRequest{
				Status:   prEntity.PRMerged,
				MergedAt: time.Now().Add(-time.Second),
			},
			expectedUpdated: false,
			expectedPr: prEntity.PullRequest{
				Status: prEntity.PRMerged,
			},
			repoError:     prErrors.ErrNotFound,
			expectedError: prErrors.ErrNotFound.Error(),
		},

		{
			what: "failed to merge pr in repo",

			prId: "pr1",
			storedPr: prEntity.PullRequest{
				Status:   prEntity.PRMerged,
				MergedAt: time.Now().Add(-time.Second),
			},
			expectedUpdated: false,
			expectedPr: prEntity.PullRequest{
				Status: prEntity.PRMerged,
			},
			repoError:     errors.New("db is down"),
			expectedError: "failed to merge pr in repo: db is down",
		},

		{
			what: "already merged",

			prId: "pr1",
			storedPr: prEntity.PullRequest{
				Status:   prEntity.PRMerged,
				MergedAt: time.Now().Add(-time.Second),
			},
			expectedUpdated: false,
			expectedPr: prEntity.PullRequest{
				Status: prEntity.PRMerged,
			},
			repoError: nil,
			noError:   true,
		},

		{
			what: "successfully merged",

			prId: "pr1",
			storedPr: prEntity.PullRequest{
				Status:   prEntity.PROpen,
				MergedAt: time.Now().Add(-time.Second),
			},
			expectedUpdated: true,
			expectedPr: prEntity.PullRequest{
				Status: prEntity.PRMerged,
			},
			repoError: nil,
			noError:   true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.
				EXPECT().
				UpdateStatus(gomock.Any(), tc.prId, gomock.Any()).
				DoAndReturn(func(
					ctx context.Context,
					pr string,
					callback interfaces.UpdateStatusHandler,
				) (prEntity.PullRequest, error) {
					updatedPr, updated := callback(tc.storedPr)

					assert.Equal(t, tc.expectedUpdated, updated)
					assert.Equal(t, tc.expectedPr.Status, updatedPr.Status)

					if updated {
						assert.NotEqual(t, tc.storedPr.MergedAt, updatedPr.MergedAt)
					} else {
						assert.Equal(t, tc.storedPr.MergedAt, updatedPr.MergedAt)
					}

					return tc.expectedPr, tc.repoError
				})

			service := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			pr, err := service.Merge(context.Background(), tc.prId)

			if tc.noError {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPr, pr)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestReassign(t *testing.T) {
	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		prId                  string
		oldReviewerId         string
		authorId              string
		teamMembers           []memberEntity.Member
		storedPr              prEntity.PullRequest
		expectedPR            prEntity.PullRequest
		repoError             error
		expectedCallbackError error
		expectedNewReviewer   string
		expectedError         string
		noError               bool
	}

	testCases := []testCase{
		{
			what: "cannot reassign",

			prId:          "pr1",
			authorId:      "u1",
			oldReviewerId: "u2",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
			},
			storedPr: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2"},
			},
			expectedCallbackError: prErrors.ErrCannotReassign,
			repoError:             prErrors.ErrCannotReassign,
			expectedError:         prErrors.ErrCannotReassign.Error(),
		},

		{
			what: "user not found",

			prId:          "pr1",
			authorId:      "u1",
			oldReviewerId: "u3",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
			},
			storedPr: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2"},
			},
			expectedCallbackError: prErrors.ErrTeamOrUserNotFound,
			repoError:             prErrors.ErrTeamOrUserNotFound,
			expectedError:         prErrors.ErrTeamOrUserNotFound.Error(),
		},

		{
			what: "pr not found",

			prId:          "pr1",
			authorId:      "u1",
			oldReviewerId: "u2",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u3",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u4",
					Activity: memberEntity.MemberActive,
				},
			},
			storedPr: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2", "u3"},
			},
			expectedCallbackError: nil,
			expectedNewReviewer:   "u4",
			repoError:             prErrors.ErrNotFound,
			expectedError:         prErrors.ErrNotFound.Error(),
		},

		{
			what: "alredy merged",

			prId:          "pr1",
			authorId:      "u1",
			oldReviewerId: "u2",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u3",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u4",
					Activity: memberEntity.MemberActive,
				},
			},
			storedPr: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PRMerged,
				Reviewers: []string{"u2", "u3"},
			},
			expectedCallbackError: prErrors.ErrAlreadyMerged,
			repoError:             prErrors.ErrAlreadyMerged,
			expectedError:         prErrors.ErrAlreadyMerged.Error(),
		},

		{
			what: "failed to reassign reviewer in repo",

			prId:          "pr1",
			authorId:      "u1",
			oldReviewerId: "u2",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u3",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u4",
					Activity: memberEntity.MemberActive,
				},
			},
			storedPr: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2", "u3"},
			},
			expectedCallbackError: nil,
			expectedNewReviewer:   "u4",
			repoError:             errors.New("db is down"),
			expectedError:         "failed to reassign reviewer in repo: db is down",
		},

		{
			what: "successfully reassign",

			prId:          "pr1",
			authorId:      "u1",
			oldReviewerId: "u2",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u3",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u4",
					Activity: memberEntity.MemberActive,
				},
			},
			storedPr: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2", "u3"},
			},
			expectedPR: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u4", "u3"},
			},
			expectedCallbackError: nil,
			expectedNewReviewer:   "u4",
			repoError:             nil,
			noError:               true,
		},

		{
			what: "ignore inactive",

			prId:          "pr1",
			authorId:      "u1",
			oldReviewerId: "u2",
			teamMembers: []memberEntity.Member{
				{
					Id:       "u1",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u2",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u3",
					Activity: memberEntity.MemberActive,
				},
				{
					Id:       "u4",
					Activity: memberEntity.MemberInactive,
				},
				{
					Id:       "u5",
					Activity: memberEntity.MemberActive,
				},
			},
			storedPr: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2", "u3"},
			},
			expectedPR: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u5", "u3"},
			},
			expectedCallbackError: nil,
			expectedNewReviewer:   "u5",
			repoError:             nil,
			noError:               true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.
				EXPECT().
				Reassign(gomock.Any(), tc.prId, tc.oldReviewerId, gomock.Any()).
				DoAndReturn(func(
					ctx context.Context,
					prId string,
					oldReviewerId string,
					callback interfaces.ReassignHandler,
				) (prEntity.PullRequest, string, error) {
					newReviewer, err := callback(tc.authorId, tc.storedPr, tc.teamMembers)

					if tc.expectedCallbackError == nil {
						assert.NoError(t, err)
						assert.Equal(t, tc.expectedNewReviewer, newReviewer)
					} else {
						assert.EqualError(t, tc.expectedCallbackError, err.Error())
					}

					return tc.expectedPR, newReviewer, tc.repoError
				})

			service := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			pr, new, err := service.Reassign(context.Background(), tc.prId, tc.oldReviewerId)

			if tc.noError {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPR, pr)
				assert.Equal(t, tc.expectedNewReviewer, new)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
