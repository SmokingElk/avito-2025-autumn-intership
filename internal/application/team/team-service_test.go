package teamservice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	teamservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/team"
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
	teamErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/interfaces"
	teamMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpsert(t *testing.T) {
	type testCase struct {
		what string

		teamName          string
		members           []memberEntity.Member
		expectedTeam      teamEntity.Team
		currentTeam       teamEntity.Team
		expectedTeamEqual bool
		repoError         error
		expectedError     string
		noError           bool
	}

	testCases := []testCase{
		{
			what: "team exists",

			teamName: "team1",
			members: []memberEntity.Member{
				{
					Id: "u1",
				},
				{
					Id: "u2",
				},
				{
					Id: "u3",
				},
			},

			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u2",
					},
					{
						Id: "u3",
					},
				},
			},

			currentTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u3",
					},
					{
						Id: "u2",
					},
				},
			},

			expectedTeamEqual: true,
			repoError:         teamErrors.ErrTeamExists,
			expectedError:     teamErrors.ErrTeamExists.Error(),
		},

		{
			what: "different members",

			teamName: "team1",
			members: []memberEntity.Member{
				{
					Id: "u1",
				},
				{
					Id: "u2",
				},
				{
					Id: "u3",
				},
			},

			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u2",
					},
					{
						Id: "u3",
					},
				},
			},

			currentTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u3",
					},
					{
						Id: "u4",
					},
				},
			},

			expectedTeamEqual: false,
			repoError:         nil,
			noError:           true,
		},

		{
			what: "different count of members",

			teamName: "team1",
			members: []memberEntity.Member{
				{
					Id: "u1",
				},
				{
					Id: "u2",
				},
				{
					Id: "u3",
				},
			},

			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u2",
					},
					{
						Id: "u3",
					},
				},
			},

			currentTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u3",
					},
				},
			},

			expectedTeamEqual: false,
			repoError:         nil,
			noError:           true,
		},

		{
			what: "different count of members",

			teamName: "team1",
			members: []memberEntity.Member{
				{
					Id: "u1",
				},
				{
					Id: "u2",
				},
				{
					Id: "u3",
				},
			},

			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u2",
					},
					{
						Id: "u3",
					},
				},
			},

			currentTeam: teamEntity.Team{},

			expectedTeamEqual: false,
			repoError:         teamErrors.ErrMemberOfOtherTeam,
			expectedError:     teamErrors.ErrMemberOfOtherTeam.Error(),
		},

		{
			what: "failed to upsert team to repo",

			teamName: "team1",
			members: []memberEntity.Member{
				{
					Id: "u1",
				},
				{
					Id: "u2",
				},
				{
					Id: "u3",
				},
			},

			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id: "u1",
					},
					{
						Id: "u2",
					},
					{
						Id: "u3",
					},
				},
			},

			currentTeam: teamEntity.Team{},

			expectedTeamEqual: false,
			repoError:         errors.New("db is down"),
			expectedError:     "failed to upsert team to repo: db is down",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTeamRepo := teamMocks.NewMockTeamRepo(ctrl)

			mockTeamRepo.
				EXPECT().
				Upsert(gomock.Any(), teamEntity.Matcher(tc.expectedTeam), gomock.Any()).
				DoAndReturn(func(ctx context.Context, team teamEntity.Team, callback interfaces.TeamMatcher) error {
					teamEqual := callback(tc.currentTeam)

					assert.Equal(t, tc.expectedTeamEqual, teamEqual)

					return tc.repoError
				})

			service := teamservice.CreateTeamService(mockTeamRepo)

			err := service.Upsert(context.Background(), tc.teamName, tc.members)

			if tc.noError {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestGetByName(t *testing.T) {
	teamName := "team1"

	type testCase struct {
		what string

		expectedTeam  teamEntity.Team
		repoError     error
		expectedError string
		noError       bool
	}

	testCases := []testCase{
		{
			what: "team not found",

			repoError:     teamErrors.ErrTeamNotFound,
			expectedError: teamErrors.ErrTeamNotFound.Error(),
		},

		{
			what: "failed to get team from repo",

			repoError:     errors.New("db is down"),
			expectedError: "failed to get team from repo: db is down",
		},

		{
			what: "successfully get team",

			noError:   true,
			repoError: nil,
			expectedTeam: teamEntity.Team{
				Id:   "team1id",
				Name: teamName,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTeamRepo := teamMocks.NewMockTeamRepo(ctrl)

			mockTeamRepo.EXPECT().GetByName(gomock.Any(), teamName).Return(tc.expectedTeam, tc.repoError)

			service := teamservice.CreateTeamService(mockTeamRepo)

			team, err := service.GetByName(context.Background(), teamName)

			if tc.noError {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTeam, team)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestDeactivateAll(t *testing.T) {
	type testCase struct {
		what string

		teamName      string
		repoError     error
		expectedError string
		noError       bool
	}

	testCases := []testCase{
		{
			what: "team not found",

			teamName:      "team1",
			repoError:     teamErrors.ErrTeamNotFound,
			expectedError: teamErrors.ErrTeamNotFound.Error(),
		},

		{
			what: "failed to deactivate all members in team",

			teamName:      "team1",
			repoError:     errors.New("db is down"),
			expectedError: "failed to deactivate in repo: db is down",
		},

		{
			what: "failed to deactivate all members in team",

			teamName:  "team1",
			repoError: nil,
			noError:   true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTeamRepo := teamMocks.NewMockTeamRepo(ctrl)

			mockTeamRepo.EXPECT().SetActivityForAll(
				gomock.Any(),
				tc.teamName,
				memberEntity.MemberInactive,
			).Return(tc.repoError)

			service := teamservice.CreateTeamService(mockTeamRepo)

			err := service.DeactivateAll(context.Background(), tc.teamName)

			if tc.noError {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
