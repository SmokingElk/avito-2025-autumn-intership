package teamhandlers_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	teamservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/team"
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
	teamErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/errors"
	teamMocks "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/mocks"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/logger"
	teamhandlers "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/handlers/team"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	log := logger.NewTest()

	type testCase struct {
		what string

		body         string
		expectedTeam teamEntity.Team
		repoError    error
		expectedCode int
		expectedBody string
	}

	testCases := []testCase{
		{
			what: "invalid body",

			body:         "{",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"BAD_REQUEST","message":"invalid body"}`,
		},

		{
			what: "team exists",

			body: `{
				"members": [
					{
						"is_active": true,
						"user_id": "u1",
						"username": "Bob"
					},
					{
						"is_active": true,
						"user_id": "u2",
						"username": "Alice"
					}
				],
				"team_name": "team1"
			}`,
			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id:       "u1",
						Username: "Bob",
						Activity: memberEntity.MemberActive,
					},
					{
						Id:       "u2",
						Username: "Alice",
						Activity: memberEntity.MemberActive,
					},
				},
			},
			repoError:    teamErrors.ErrTeamExists,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"TEAM_EXISTS","message":"team_name already exists"}`,
		},

		{
			what: "member of other team",

			body: `{
				"members": [
					{
						"is_active": true,
						"user_id": "u1",
						"username": "Bob"
					},
					{
						"is_active": true,
						"user_id": "u2",
						"username": "Alice"
					}
				],
				"team_name": "team1"
			}`,
			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id:       "u1",
						Username: "Bob",
						Activity: memberEntity.MemberActive,
					},
					{
						Id:       "u2",
						Username: "Alice",
						Activity: memberEntity.MemberActive,
					},
				},
			},
			repoError:    teamErrors.ErrMemberOfOtherTeam,
			expectedCode: http.StatusConflict,
			expectedBody: `{"code":"MEMBER_OF_OTHER_TEAM","message":"User is member of other team"}`,
		},

		{
			what: "failed to create team",

			body: `{
				"members": [
					{
						"is_active": true,
						"user_id": "u1",
						"username": "Bob"
					},
					{
						"is_active": true,
						"user_id": "u2",
						"username": "Alice"
					}
				],
				"team_name": "team1"
			}`,
			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id:       "u1",
						Username: "Bob",
						Activity: memberEntity.MemberActive,
					},
					{
						Id:       "u2",
						Username: "Alice",
						Activity: memberEntity.MemberActive,
					},
				},
			},
			repoError:    errors.New("db is down"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"INTERNAL_SERVER_ERROR","message":"failed to create team: ` +
				`failed to upsert team to repo: db is down"}`,
		},

		{
			what: "failed to create team",

			body: `{
				"members": [
					{
						"is_active": true,
						"user_id": "u1",
						"username": "Bob"
					},
					{
						"is_active": true,
						"user_id": "u2",
						"username": "Alice"
					}
				],
				"team_name": "team1"
			}`,
			expectedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id:       "u1",
						Username: "Bob",
						Activity: memberEntity.MemberActive,
					},
					{
						Id:       "u2",
						Username: "Alice",
						Activity: memberEntity.MemberActive,
					},
				},
			},
			repoError:    nil,
			expectedCode: http.StatusCreated,
			expectedBody: `{"team":{"team_name":"team1","members":[{"user_id":"u1","username":"Bob","is_active":true},` +
				`{"user_id":"u2","username":"Alice","is_active":true}]}}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			teamRepo := teamMocks.NewMockTeamRepo(ctrl)

			teamRepo.EXPECT().Upsert(
				gomock.Any(),
				teamEntity.Matcher(tc.expectedTeam),
				gomock.Any(),
			).Return(tc.repoError).MaxTimes(1)

			teamService := teamservice.CreateTeamService(teamRepo)

			handlers := teamhandlers.CreateTeamHandlers(teamService, log)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/", handlers.Add)

			body := bytes.NewBufferString(tc.body)
			req := httptest.NewRequest("POST", "/", body)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}

func TestGet(t *testing.T) {
	log := logger.NewTest()

	type testCase struct {
		what string

		teamName     string
		storedTeam   teamEntity.Team
		repoError    error
		expectedCode int
		expectedBody string
	}

	testCases := []testCase{
		{
			what: "invalid team_name param",

			teamName:     "",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"BAD_REQUEST","message":"invalid team_name param"}`,
		},

		{
			what: "team not found",

			teamName:     "team1",
			repoError:    teamErrors.ErrTeamNotFound,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":"NOT_FOUND","message":"resource not found"}`,
		},

		{
			what: "failed to get team",

			teamName:     "team1",
			repoError:    errors.New("db is down"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"INTERNAL_SERVER_ERROR","message":"failed to get team: ` +
				`failed to get team from repo: db is down"}`,
		},

		{
			what: "failed to get team",

			teamName: "team1",
			storedTeam: teamEntity.Team{
				Name: "team1",
				Members: []memberEntity.Member{
					{
						Id:       "u1",
						Username: "Bob",
						Activity: memberEntity.MemberActive,
					},
					{
						Id:       "u2",
						Username: "Alice",
						Activity: memberEntity.MemberActive,
					},
				},
			},
			repoError:    nil,
			expectedCode: http.StatusOK,
			expectedBody: `{"team_name":"team1","members":[{"user_id":"u1","username":"Bob","is_active":true},` +
				`{"user_id":"u2","username":"Alice","is_active":true}]}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tc.what), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			teamRepo := teamMocks.NewMockTeamRepo(ctrl)

			teamRepo.EXPECT().GetByName(
				gomock.Any(),
				tc.teamName,
			).Return(tc.storedTeam, tc.repoError).MaxTimes(1)

			teamService := teamservice.CreateTeamService(teamRepo)

			handlers := teamhandlers.CreateTeamHandlers(teamService, log)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/", handlers.Get)

			req := httptest.NewRequest("GET", fmt.Sprintf("/?team_name=%s", tc.teamName), nil)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedCode, recorder.Code)
			assert.Equal(t, tc.expectedBody, recorder.Body.String())
		})
	}
}
