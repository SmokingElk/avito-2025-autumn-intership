package memberhandlers_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	memberservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/member"
	pullrequestservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/pull-request"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	memberErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/errors"
	memberMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/mocks"
	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
	prMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/mocks"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/logger"
	memberhandlers "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/handlers/member"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSetIsActive(t *testing.T) {
	log := logger.NewTest()

	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		userId           string
		body             string
		expectedActivity memberEntity.MemberActivity
		expectedMember   memberEntity.Member
		repoError        error
		expectedCode     int
		expectedBody     string
	}

	testCases := []testCase{
		{
			what: "invalid body",

			body:         "{",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":{"code":"BAD_REQUEST","message":"invalid body"}}`,
		},

		{
			what: "user not found",

			userId: "u99",
			body: `{
				"is_active": false,
  				"user_id": "u99"
			}`,
			expectedActivity: memberEntity.MemberInactive,
			repoError:        memberErrors.ErrMemberNotFound,
			expectedCode:     http.StatusNotFound,
			expectedBody:     `{"error":{"code":"NOT_FOUND","message":"resource not found"}}`,
		},

		{
			what: "failed to set is active",

			userId: "u1",
			body: `{
				"is_active": false,
  				"user_id": "u1"
			}`,
			expectedActivity: memberEntity.MemberInactive,
			repoError:        errors.New("db is down"),
			expectedCode:     http.StatusInternalServerError,
			expectedBody: `{"error":{"code":"INTERNAL_SERVER_ERROR","message":"failed to set is active: ` +
				`failed to set active in repo: db is down"}}`,
		},

		{
			what: "successfully set inactive",

			userId: "u1",
			body: `{
				"is_active": false,
  				"user_id": "u1"
			}`,
			expectedActivity: memberEntity.MemberInactive,
			expectedMember: memberEntity.Member{
				Id:       "u1",
				Username: "Bob",
				Activity: memberEntity.MemberInactive,
				TeamName: "team1",
			},
			repoError:    nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"user_id":"u1","username":"Bob","team_name":"team1","is_active":false}`,
		},

		{
			what: "successfully set active",

			userId: "u1",
			body: `{
				"is_active": true,
  				"user_id": "u1"
			}`,
			expectedActivity: memberEntity.MemberActive,
			expectedMember: memberEntity.Member{
				Id:       "u1",
				Username: "Bob",
				Activity: memberEntity.MemberActive,
				TeamName: "team1",
			},
			repoError:    nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"user_id":"u1","username":"Bob","team_name":"team1","is_active":true}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemberRepo := memberMocks.NewMockMemberRepo(ctrl)

			mockMemberRepo.EXPECT().SetActivity(
				gomock.Any(),
				tc.userId,
				tc.expectedActivity,
			).Return(tc.expectedMember, tc.repoError).MaxTimes(1)

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			memberService := memberservice.CreateMemberService(mockMemberRepo)
			pullRequestService := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			handlers := memberhandlers.CreateMemberHandlers(memberService, pullRequestService, log)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", handlers.SetIsActive)

			body := bytes.NewBufferString(tc.body)
			req := httptest.NewRequest("POST", "/", body)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}

func TestGetReview(t *testing.T) {
	log := logger.NewTest()

	config := config.PullRequestConfig{
		OutLimit:             10,
		TargetReviewersCount: 2,
	}

	type testCase struct {
		what string

		userId       string
		expectedPRs  []prEntity.PullRequest
		repoError    error
		expectedCode int
		expectedBody string
	}

	testCases := []testCase{
		{
			what: "invalid user_id param",

			userId:       "",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":{"code":"BAD_REQUEST","message":"invalid user_id param"}}`,
		},

		{
			what: "failed to get PRs by user id",

			userId:       "u1",
			repoError:    errors.New("db is down"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":{"code":"INTERNAL_SERVER_ERROR","message":"failed to get pr's by user id: ` +
				`failed to get pull requests from repo: db is down"}}`,
		},

		{
			what: "successfully get PRs",

			userId: "u1",
			expectedPRs: []prEntity.PullRequest{
				{
					Id:       "pr1",
					Name:     "pull request 1",
					AuthorId: "u2",
					Status:   prEntity.PROpen,
				},
				{
					Id:       "pr2",
					Name:     "pull request 2",
					AuthorId: "u3",
					Status:   prEntity.PRMerged,
				},
			},
			repoError:    nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"user_id":"u1","pull_requests":[{"pull_request_id":"pr1","pull_request_name":"pull request 1",` +
				`"author_id":"u2","status":"OPEN"},{"pull_request_id":"pr2","pull_request_name":"pull request 2",` +
				`"author_id":"u3","status":"MERGED"}]}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemberRepo := memberMocks.NewMockMemberRepo(ctrl)

			mockPullRequestRepo := prMocks.NewMockPullRequestRepo(ctrl)

			mockPullRequestRepo.EXPECT().GetByReviewer(
				gomock.Any(),
				tc.userId,
				config.OutLimit,
			).Return(tc.expectedPRs, tc.repoError).MaxTimes(1)

			memberService := memberservice.CreateMemberService(mockMemberRepo)
			pullRequestService := pullrequestservice.CreatePullRequestService(mockPullRequestRepo, &config)

			handlers := memberhandlers.CreateMemberHandlers(memberService, pullRequestService, log)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/", handlers.GetReview)

			req := httptest.NewRequest("GET", fmt.Sprintf("/?user_id=%s", tc.userId), nil)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}
