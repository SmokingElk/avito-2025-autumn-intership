package statsservice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	statsservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/statistics"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/entity"
	statsMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAssignmentsPerMember(t *testing.T) {
	type testCase struct {
		what string

		limit              int
		offset             int
		expectedStatistics []entity.AssignmentsPerMember
		repoError          error
		expectedError      string
		noError            bool
	}

	testCases := []testCase{
		{
			what: "failed to get assignments stats from repo",

			limit:         5,
			offset:        10,
			repoError:     errors.New("db is down"),
			expectedError: "failed to get assignments stats from repo: db is down",
		},

		{
			what: "successfully get statistics",

			limit:     2,
			offset:    10,
			repoError: nil,
			noError:   true,
			expectedStatistics: []entity.AssignmentsPerMember{
				{
					MemberId:         "u1",
					AssignmentsCount: 5,
				},
				{
					MemberId:         "u2",
					AssignmentsCount: 2,
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStatsRepo := statsMocks.NewMockStatsRepo(ctrl)

			mockStatsRepo.EXPECT().GetAssignmentsPerMember(
				gomock.Any(),
				tc.limit,
				tc.offset,
			).Return(tc.expectedStatistics, tc.repoError)

			statsService := statsservice.CreateStatsService(mockStatsRepo)

			stats, err := statsService.GetAssignmentsPerMember(context.Background(), tc.limit, tc.offset)

			if tc.noError {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedStatistics, stats)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
