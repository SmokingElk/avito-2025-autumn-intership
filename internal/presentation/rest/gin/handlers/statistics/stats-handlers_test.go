package statshandlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	statsservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/statistics"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/entity"
	statsMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/mocks"
	statshandlers "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/handlers/statistics"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAssignmentsPerMember(t *testing.T) {
	type testCase struct {
		what string

		limitParam    string
		offsetParam   string
		limit         int
		offset        int
		repoError     error
		expectedStats []entity.AssignmentsPerMember
		expectedCode  int
		expectedBody  string
	}

	testCases := []testCase{
		{
			what: "invalid limit param",

			limitParam:   "",
			offsetParam:  "1",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"BAD_REQUEST","message":"invalid limit param"}`,
		},

		{
			what: "invalid offset param",

			limitParam:   "1",
			offsetParam:  "",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"BAD_REQUEST","message":"invalid offset param"}`,
		},

		{
			what: "failed to get assignmnets per member",

			limitParam:   "2",
			offsetParam:  "5",
			limit:        2,
			offset:       5,
			repoError:    errors.New("db is down"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"INTERNAL_SERVER_ERROR","message":"failed to get assignmnets per member: ` +
				`failed to get assignments stats from repo: db is down"}`,
		},

		{
			what: "failed to get assignmnets per member",

			limitParam:  "2",
			offsetParam: "5",
			limit:       2,
			offset:      5,
			repoError:   nil,
			expectedStats: []entity.AssignmentsPerMember{
				{
					MemberId:         "u1",
					AssignmentsCount: 5,
				},
				{
					MemberId:         "u2",
					AssignmentsCount: 2,
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"count":2,"results":[{"member_id":"u1","assigments_count":5},` +
				`{"member_id":"u2","assigments_count":2}]}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			statsRepo := statsMocks.NewMockStatsRepo(ctrl)

			statsRepo.EXPECT().GetAssignmentsPerMember(
				gomock.Any(),
				tc.limit,
				tc.offset,
			).Return(tc.expectedStats, tc.repoError).MaxTimes(1)

			statsService := statsservice.CreateStatsService(statsRepo)

			handlers := statshandlers.CreateStatsHandlers(statsService)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/", handlers.GetAssignmentsPerMember)

			req := httptest.NewRequest("GET", fmt.Sprintf("/?limit=%s&offset=%s", tc.limitParam, tc.offsetParam), nil)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}
