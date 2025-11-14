package memberhandlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/SmokingElk/avito-2025-autumn-intership/docs"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	memberErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/errors"
	memberInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/interfaces"
	pullRequestInterfaces "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/auth"
	request_id "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/request-id"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type MemberHandlers struct {
	memberService      memberInterfaces.MemberService
	pullRequestService pullRequestInterfaces.PullRequestService
	logger             zerolog.Logger
}

func CreateMemberHandlers(
	memberService memberInterfaces.MemberService,
	pullRequestService pullRequestInterfaces.PullRequestService,
	log zerolog.Logger,
) *MemberHandlers {
	return &MemberHandlers{
		memberService:      memberService,
		pullRequestService: pullRequestService,
		logger:             log,
	}
}

// Add godoc
// @Summary Установить флаг активности пользователя
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body docs.SetIsActiveRequest true "Данные для обновления"
// @Success 200 {object} docs.SetIsActiveResponse "Обновленный пользователь"
// @Failure 401 {object} docs.ErrorResponse "Нет/неверный админский токен"
// @Failure 404 {object} docs.ErrorResponse "Пользователь не найден"
// @Router /users/setIsActive [post]
func (h *MemberHandlers) SetIsActive(ctx *gin.Context) {
	log := h.localLogger(ctx, "SetIsActive")

	var request docs.SetIsActiveRequest

	if err := ctx.BindJSON(&request); err != nil {
		log.Warn().Msg("invalid body")
		ctx.Status(http.StatusBadRequest)
		return
	}

	member, err := h.memberService.SetIsActive(ctx.Request.Context(), request.UserId, request.IsActive)

	if err != nil {
		switch {
		case errors.Is(err, memberErrors.ErrMemberNotFound):
			log.Warn().Msg("user not found")
			ctx.AbortWithStatusJSON(http.StatusNotFound, docs.NewErrorResponse(
				"NOT_FOUND",
				"resource not found",
			))
		default:
			log.Error().Err(err).Msg("failed to set is active")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
				"INTERNAL_SERVER_ERROR",
				fmt.Sprintf("failed to set is active: %s", err.Error()),
			))
		}

		return
	}

	resp := docs.ToSetIsActiveResponse(member)
	ctx.JSON(http.StatusOK, resp)

	log.Info().Msg("successfully updated member activity")
}

// Add godoc
// @Summary Получить PR'ы, где пользователь установлен ревьювером
// @Tags Users
// @Security BearerAuth
// @Param user_id query string true "Идентификатор пользователя"
// @Produce json
// @Success 200 {object} docs.GetReviewResponse "Список PR'ов пользователя"
// @Failure 401 {object} docs.ErrorResponse "Нет/неверный админский токен"
// @Router /users/getReview [get]
func (h *MemberHandlers) GetReview(ctx *gin.Context) {
	log := h.localLogger(ctx, "GetReview")

	userId := ctx.Query("user_id")

	if userId == "" {
		log.Warn().Msg("invalid user_id param")
		ctx.Status(http.StatusBadRequest)
		return
	}

	prs, err := h.pullRequestService.GetByReviewer(ctx.Request.Context(), userId)

	if err != nil {
		log.Error().Err(err).Msg("failed to get pr's by user id")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
			"INTERNAL_SERVER_ERROR",
			fmt.Sprintf("failed to get pr's by user id: %s", err.Error()),
		))

		return
	}

	resp := docs.GetReviewResponse{
		UserId:       userId,
		PullRequests: make([]docs.GetReviewPRResponse, 0, len(prs)),
	}

	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, docs.ToReviewPRResponse(pr))
	}

	ctx.JSON(http.StatusOK, resp)

	log.Info().Msg("successfully get members")
}

func (h *MemberHandlers) localLogger(ctx *gin.Context, opName string) zerolog.Logger {
	log := h.logger.With().
		Str("op", opName).
		Str("requestId", ctx.GetString(request_id.REQUEST_ID_PARAM)).
		Logger()

	return log
}

func InitMemberHandlers(
	r *gin.RouterGroup,
	log zerolog.Logger,
	memberService memberInterfaces.MemberService,
	pullRequestService pullRequestInterfaces.PullRequestService,
	cfg *config.RestConfig,
) {
	handlers := CreateMemberHandlers(memberService, pullRequestService, log)

	group := r.Group("users")

	{
		group.POST("setIsActive", auth.WithAuth(cfg), handlers.SetIsActive)
		group.GET("getReview", auth.WithAuth(cfg), handlers.GetReview)
	}
}
