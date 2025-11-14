package di

import (
	memberservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/member"
	pullrequestservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/pull-request"
	teamservice "github.com/SmokingElk/avito-2025-autumn-intership/internal/application/team"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/clients/postgres"
	memberrepopg "github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/member"
	pullrequestrepopg "github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/pull-request"
	teamrepopg "github.com/SmokingElk/avito-2025-autumn-intership/internal/infrastructure/repos/team"
	rest "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func MustConfigureApp(r *gin.Engine, cfg *config.Config, log zerolog.Logger) func() {
	conn, err := postgres.CreateConnection(&cfg.PostgresConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}

	memberRepo := memberrepopg.CreateMemberRepoPg(conn)
	teamRepo := teamrepopg.CreateTeamRepoPg(conn)
	pullRequestRepo := pullrequestrepopg.CreatePullRequestRepoPg(conn)

	memberService := memberservice.CreateMemberService(memberRepo)
	teamService := teamservice.CreateTeamService(teamRepo)
	pullrequestservice := pullrequestservice.CreatePullRequestService(pullRequestRepo, &cfg.PullRequestConfig)

	rest.InitRoutes(r, &cfg.RestConfig, log, memberService, teamService, pullrequestservice)

	return func() {
		conn.Close()
	}
}
