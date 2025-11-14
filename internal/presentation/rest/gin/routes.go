package rest

import (
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	memberInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/interfaces"
	pullRequestInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
	teamInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/interfaces"
	memberhandlers "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/handlers/member"
	teamhandlers "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/handlers/team"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/cors"
	request_id "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/request-id"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/SmokingElk/avito-2025-autumn-intership/docs"
)

// @title PR Reviewer Assignment Service (Test Task, Fall 2025)
// @version 1.0.0
// @description API for pull requests managment in teams
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func InitRoutes(
	r *gin.Engine,
	cfg *config.RestConfig,
	log zerolog.Logger,
	memberService memberInterfaces.MemberService,
	teamService teamInterfaces.TeamService,
	pullRequestService pullRequestInterfaces.PullRequestService,
) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("api/v1")

	api.Use(cors.CORS(cfg.AllowOrigin))
	api.Use(request_id.AddRequestId())

	memberhandlers.InitMemberHandlers(api, log, memberService, pullRequestService, cfg)
	teamhandlers.InitTeamHandlers(api, log, teamService, cfg)
}
