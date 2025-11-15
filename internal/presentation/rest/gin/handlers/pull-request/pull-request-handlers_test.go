package pullrequesthandlers_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pullrequestservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/pull-request"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
	prErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/errors"
	prMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/mocks"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/logger"
	pullrequesthandlers "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/handlers/pull-request"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	log := logger.NewTest()

	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		body                    string
		expectedPR              prEntity.PullRequest
		expectedPRWithReviewers prEntity.PullRequest
		repoError               error
		expectedCode            int
		expectedBody            string
	}

	testCases := []testCase{
		{
			what: "invalid body",

			body: `{
				"author_id": "u1",
  				"pull_request_id": "pr1",
  				"pull_request_name": "pull request 1"
  			`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"BAD_REQUEST","message":"invalid body"}`,
		},

		{
			what: "user not found",

			body: `{
				"author_id": "u1",
  				"pull_request_id": "pr1",
  				"pull_request_name": "pull request 1"
  			}`,
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			repoError:    prErrors.ErrTeamOrUserNotFound,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":"NOT_FOUND","message":"resource not found"}`,
		},

		{
			what: "pr exists",

			body: `{
				"author_id": "u1",
  				"pull_request_id": "pr1",
  				"pull_request_name": "pull request 1"
  			}`,
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			repoError:    prErrors.ErrAlreadyExists,
			expectedCode: http.StatusConflict,
			expectedBody: `{"code":"PR_EXISTS","message":"PR id already exists"}`,
		},

		{
			what: "failed to create pr",

			body: `{
				"author_id": "u1",
  				"pull_request_id": "pr1",
  				"pull_request_name": "pull request 1"
  			}`,
			expectedPR: prEntity.PullRequest{
				Id:       "pr1",
				Name:     "pull request 1",
				AuthorId: "u1",
				Status:   prEntity.PROpen,
			},
			repoError:    errors.New("db is down"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"INTERNAL_SERVER_ERROR","message":"failed to create pr: failed to store pr in repo: ` +
				`db is down"}`,
		},

		{
			what: "successfully create pull request",

			body: `{
				"author_id": "u1",
  				"pull_request_id": "pr1",
  				"pull_request_name": "pull request 1"
  			}`,
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
			repoError:    nil,
			expectedCode: http.StatusCreated,
			expectedBody: `{"pr":{"pull_request_id":"pr1","pull_request_name":"pull request 1","author_id":"u1",` +
				`"status":"OPEN","assigned_reviewers":["u2","u3"]}}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.EXPECT().Create(
				gomock.Any(),
				prEntity.Matcher(tc.expectedPR),
				gomock.Any(),
			).Return(tc.expectedPRWithReviewers, tc.repoError).MaxTimes(1)

			pullRequestService := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			handlers := pullrequesthandlers.CreatePullRequestHandlers(pullRequestService, log)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", handlers.Create)

			body := bytes.NewBufferString(tc.body)
			req := httptest.NewRequest("POST", "/", body)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}

func TestMerge(t *testing.T) {
	log := logger.NewTest()

	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		prId         string
		body         string
		updatedPR    prEntity.PullRequest
		repoError    error
		expectedCode int
		expectedBody string
	}

	testCases := []testCase{
		{
			what: "invalid body",

			body: `{
				"pull_request_id": "pr1"
			`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"BAD_REQUEST","message":"invalid body"}`,
		},

		{
			what: "pr not found",

			body: `{
				"pull_request_id": "pr1"
			}`,
			prId:         "pr1",
			repoError:    prErrors.ErrNotFound,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":"NOT_FOUND","message":"resource not found"}`,
		},

		{
			what: "failed to merge pr",

			body: `{
				"pull_request_id": "pr1"
			}`,
			prId:         "pr1",
			repoError:    errors.New("db is down"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"INTERNAL_SERVER_ERROR","message":"failed to merge pr: ` +
				`failed to merge pr in repo: db is down"}`,
		},

		{
			what: "successfully merged",

			body: `{
				"pull_request_id": "pr1"
			}`,
			prId:      "pr1",
			repoError: nil,
			updatedPR: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u1",
				Status:    prEntity.PRMerged,
				Reviewers: []string{"u2", "u3"},
				MergedAt:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"pr":{"pull_request_id":"pr1","pull_request_name":"pull request 1","author_id":"u1",` +
				`"status":"MERGED","assigned_reviewers":["u2","u3"],"mergedAt":"1970-01-01T00:00:00Z"}}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.EXPECT().UpdateStatus(
				gomock.Any(),
				tc.prId,
				gomock.Any(),
			).Return(tc.updatedPR, tc.repoError).MaxTimes(1)

			pullRequestService := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			handlers := pullrequesthandlers.CreatePullRequestHandlers(pullRequestService, log)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", handlers.Merge)

			body := bytes.NewBufferString(tc.body)
			req := httptest.NewRequest("POST", "/", body)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}

func TestReassign(t *testing.T) {
	log := logger.NewTest()

	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		prId          string
		oldReviewerId string
		replacedBy    string
		body          string
		updatedPR     prEntity.PullRequest
		repoError     error
		expectedCode  int
		expectedBody  string
	}

	testCases := []testCase{
		{
			what: "invalid body",

			body: `{
				"old_reviewer_id": "u1",
				"pull_request_id": "pr1"
			`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"BAD_REQUEST","message":"invalid body"}`,
		},

		{
			what: "pr not found",

			body: `{
				"old_reviewer_id": "u1",
				"pull_request_id": "pr1"
			}`,
			prId:          "pr1",
			oldReviewerId: "u1",
			repoError:     prErrors.ErrNotFound,
			expectedCode:  http.StatusNotFound,
			expectedBody:  `{"code":"NOT_FOUND","message":"resource not found"}`,
		},

		{
			what: "user not found",

			body: `{
				"old_reviewer_id": "u1",
				"pull_request_id": "pr1"
			}`,
			prId:          "pr1",
			oldReviewerId: "u1",
			repoError:     prErrors.ErrTeamOrUserNotFound,
			expectedCode:  http.StatusNotFound,
			expectedBody:  `{"code":"NOT_FOUND","message":"resource not found"}`,
		},

		{
			what: "pr already merged",

			body: `{
				"old_reviewer_id": "u1",
				"pull_request_id": "pr1"
			}`,
			prId:          "pr1",
			oldReviewerId: "u1",
			repoError:     prErrors.ErrAlreadyMerged,
			expectedCode:  http.StatusConflict,
			expectedBody:  `{"code":"NOT_FOUND","message":"resource not found"}`,
		},

		{
			what: "failed to reassign",

			body: `{
				"old_reviewer_id": "u1",
				"pull_request_id": "pr1"
			}`,
			prId:          "pr1",
			oldReviewerId: "u1",
			repoError:     errors.New("db is down"),
			expectedCode:  http.StatusInternalServerError,
			expectedBody: `{"code":"INTERNAL_SERVER_ERROR","message":"failed to reassign: ` +
				`failed to reassign reviewer in repo: db is down"}`,
		},

		{
			what: "failed to reassign",

			body: `{
				"old_reviewer_id": "u1",
				"pull_request_id": "pr1"
			}`,
			prId:          "pr1",
			oldReviewerId: "u1",
			replacedBy:    "u2",
			repoError:     nil,
			updatedPR: prEntity.PullRequest{
				Id:        "pr1",
				Name:      "pull request 1",
				AuthorId:  "u3",
				Status:    prEntity.PROpen,
				Reviewers: []string{"u2", "u4"},
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"pr":{"pull_request_id":"pr1","pull_request_name":"pull request 1","author_id":"u3",` +
				`"status":"OPEN","assigned_reviewers":["u2","u4"]},"replaced_by":"u2"}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.EXPECT().Reassign(
				gomock.Any(),
				tc.prId,
				tc.oldReviewerId,
				gomock.Any(),
			).Return(tc.updatedPR, tc.replacedBy, tc.repoError).MaxTimes(1)

			pullRequestService := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			handlers := pullrequesthandlers.CreatePullRequestHandlers(pullRequestService, log)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", handlers.Reassign)

			body := bytes.NewBufferString(tc.body)
			req := httptest.NewRequest("POST", "/", body)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}
