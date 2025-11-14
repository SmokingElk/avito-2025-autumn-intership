package rest

import (
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	memberInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/interfaces"
	pullRequestInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
	teamInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/middleware/cors"
	request_id "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/middleware/request-id"
	"github.com/gin-gonic/gin"
)

func InitRoutes(
	r *gin.Engine,
	cfg *config.RestConfig,
	memberService memberInterfaces.MemberService,
	teamService teamInterfaces.TeamService,
	pullRequestService pullRequestInterfaces.PullRequestService,
) {
	api := r.Group("api/v1")

	api.Use(cors.CORS(cfg.AllowOrigin))
	api.Use(request_id.AddRequestId())

}
