package memberservice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	memberservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/member"
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	memberErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/errors"
	memberMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSetIsActive(t *testing.T) {
	userId := "u1"

	type testCase struct {
		what     string
		isActive bool

		expectedActivity memberEntity.MemberActivity
		expectedMember   memberEntity.Member
		repoError        error
		expectedError    string
		noError          bool
	}

	testCases := []testCase{
		{
			what: "member not found",

			isActive:         false,
			expectedActivity: memberEntity.MemberInactive,
			repoError:        memberErrors.ErrMemberNotFound,
			expectedError:    memberErrors.ErrMemberNotFound.Error(),
		},

		{
			what: "failed to set active in repo",

			isActive:         false,
			expectedActivity: memberEntity.MemberInactive,
			repoError:        errors.New("db is down"),
			expectedError:    "failed to set active in repo: db is down",
		},

		{
			what: "successfully set inactive",

			isActive:         false,
			expectedActivity: memberEntity.MemberInactive,
			expectedMember: memberEntity.Member{
				Id:       userId,
				Activity: memberEntity.MemberInactive,
			},
			repoError: nil,
			noError:   true,
		},

		{
			what: "successfully set active",

			isActive:         true,
			expectedActivity: memberEntity.MemberActive,
			expectedMember: memberEntity.Member{
				Id:       userId,
				Activity: memberEntity.MemberActive,
			},
			repoError: nil,
			noError:   true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMemberRepo := memberMocks.NewMockMemberRepo(ctrl)

			mockMemberRepo.EXPECT().SetActivity(
				gomock.Any(),
				userId,
				tc.expectedActivity,
			).Return(tc.expectedMember, tc.repoError)

			service := memberservice.CreateMemberService(mockMemberRepo)

			member, err := service.SetIsActive(context.Background(), userId, tc.isActive)

			if tc.noError {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMember, member)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
