package pullrequesthandlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/SmokingElk/avito-2025-autumn-intership/docs"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	prErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/auth"
	request_id "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/request-id"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type PullRequestHandlers struct {
	pullRequestService interfaces.PullRequestService
	logger             zerolog.Logger
}

func CreatePullRequestHandlers(pullRequestService interfaces.PullRequestService, log zerolog.Logger) *PullRequestHandlers {
	return &PullRequestHandlers{
		pullRequestService: pullRequestService,
		logger:             log,
	}
}

// Add godoc
// @Summary Создать PR и автоматически назначить до 2 ревьюверов из команды авторы
// @Tags PullRequests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body docs.CreatePRRequest true "Данные для создания"
// @Success 201 {object} docs.CreatePRResponse "PR создан"
// @Failure 404 {object} docs.ErrorResponse "Автор/команда не найдены"
// @Failure 409 {object} docs.ErrorResponse "PR уже существует"
// @Router /pullRequest/create [post]
func (h *PullRequestHandlers) Create(ctx *gin.Context) {
	log := h.localLogger(ctx, "Create")

	var request docs.CreatePRRequest

	if err := ctx.BindJSON(&request); err != nil {
		log.Warn().Msg("invalid body")
		ctx.Status(http.StatusBadRequest)
		return
	}

	pr, err := h.pullRequestService.Create(ctx.Request.Context(), request.Id, request.Name, request.AuthorId)

	if err != nil {
		switch {
		case errors.Is(err, prErrors.ErrTeamOrUserNotFound):
			log.Warn().Msg("team or user not found")
			ctx.AbortWithStatusJSON(http.StatusNotFound, docs.NewErrorResponse(
				"NOT_FOUND",
				"resource not found",
			))

		case errors.Is(err, prErrors.ErrAlreadyExists):
			log.Warn().Msg("pr already exists")
			ctx.AbortWithStatusJSON(http.StatusNotFound, docs.NewErrorResponse(
				"PR_EXISTS",
				"PR id already exists",
			))

		default:
			log.Error().Err(err).Msg("failed to create pr")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
				"INTERNAL_SERVER_ERROR",
				fmt.Sprintf("failed to create pr: %s", err.Error()),
			))
		}

		return
	}

	resp := docs.CreatePRResponse{
		Pr: docs.PRResponseObject{
			Id:                pr.Id,
			Name:              pr.Name,
			AuthorId:          pr.AuthorId,
			Status:            string(pr.Status),
			AssignedReviewers: pr.Reviewers,
		},
	}

	ctx.JSON(http.StatusCreated, resp)

	log.Info().Msg("successfully created pr")
}

// Add godoc
// @Summary Пометить PR как MERGED (идемпотентная операция)
// @Tags PullRequests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body docs.MergePRRequest true "Идентификатор PR"
// @Success 200 {object} docs.MergePRResponse "PR в состоянии MERGED"
// @Failure 404 {object} docs.ErrorResponse "PR не найден"
// @Router /pullRequest/merge [post]
func (h *PullRequestHandlers) Merge(ctx *gin.Context) {
	log := h.localLogger(ctx, "Merge")

	var request docs.MergePRRequest

	if err := ctx.BindJSON(&request); err != nil {
		log.Warn().Msg("invalid body")
		ctx.Status(http.StatusBadRequest)
		return
	}

	mergedPr, err := h.pullRequestService.Merge(ctx.Request.Context(), request.Id)

	if err != nil {
		switch {
		case errors.Is(err, prErrors.ErrNotFound):
			log.Warn().Msg("pr not found")
			ctx.AbortWithStatusJSON(http.StatusNotFound, docs.NewErrorResponse(
				"NOT_FOUND",
				"resource not found",
			))

		default:
			log.Error().Err(err).Msg("failed to create pr")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
				"INTERNAL_SERVER_ERROR",
				fmt.Sprintf("failed to create pr: %s", err.Error()),
			))
		}

		return
	}

	resp := docs.MergePRResponse{
		Pr: docs.MergePRResponseObject{
			Id:                mergedPr.Id,
			Name:              mergedPr.Name,
			AuthorId:          mergedPr.AuthorId,
			Status:            string(mergedPr.Status),
			AssignedReviewers: mergedPr.Reviewers,
			MergedAt:          mergedPr.MergedAt,
		},
	}

	ctx.JSON(http.StatusCreated, resp)

	log.Info().Msg("successfully merged pr")
}

// Add godoc
// @Summary Переназначить конкретного ревьювера на другого из его команды
// @Tags PullRequests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body docs.ReassignRequest true "Данные для переназначения"
// @Success 200 {object} docs.ReassignResponse "Переназначение выполнено"
// @Failure 404 {object} docs.ErrorResponse "PR или пользователь найден"
// @Failure 409 {object} docs.ErrorResponse "Нарушение доменных правил переназначения"
// @Router /pullRequest/reassign [post]
func (h *PullRequestHandlers) Reassign(ctx *gin.Context) {
	log := h.localLogger(ctx, "Reassign")

	var request docs.ReassignRequest

	if err := ctx.BindJSON(&request); err != nil {
		log.Warn().Msg("invalid body")
		ctx.Status(http.StatusBadRequest)
		return
	}

	pr, new, err := h.pullRequestService.Reassign(ctx.Request.Context(), request.Id, request.OldReviewerId)

	if err != nil {
		switch {
		case errors.Is(err, prErrors.ErrNotFound):
			log.Warn().Msg("pr not found")
			ctx.AbortWithStatusJSON(http.StatusNotFound, docs.NewErrorResponse(
				"NOT_FOUND",
				"resource not found",
			))

		case errors.Is(err, prErrors.ErrAlreadyMerged):
			log.Warn().Msg("pr already merged")
			ctx.AbortWithStatusJSON(http.StatusConflict, docs.NewErrorResponse(
				"NOT_FOUND",
				"resource not found",
			))

		default:
			log.Error().Err(err).Msg("failed to create pr")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
				"INTERNAL_SERVER_ERROR",
				fmt.Sprintf("failed to create pr: %s", err.Error()),
			))
		}

		return
	}

	resp := docs.ReassignResponse{
		Pr: docs.PRResponseObject{
			Id:                pr.Id,
			Name:              pr.Name,
			AuthorId:          pr.AuthorId,
			Status:            string(pr.Status),
			AssignedReviewers: pr.Reviewers,
		},
		ReplacedBy: new,
	}

	ctx.JSON(http.StatusCreated, resp)

	log.Info().Msg("successfully reassigned")
}

func (h *PullRequestHandlers) localLogger(ctx *gin.Context, opName string) zerolog.Logger {
	log := h.logger.With().
		Str("op", opName).
		Str("requestId", ctx.GetString(request_id.REQUEST_ID_PARAM)).
		Logger()

	return log
}

func InitPullRequestHandlers(
	r *gin.RouterGroup,
	log zerolog.Logger,
	pullRequestService interfaces.PullRequestService,
	cfg *config.RestConfig,
) {
	h := CreatePullRequestHandlers(pullRequestService, log)

	group := r.Group("pullRequest")

	{
		group.POST("create", auth.WithAuth(cfg), h.Create)
		group.POST("merge", auth.WithAuth(cfg), h.Merge)
		group.POST("reassign", auth.WithAuth(cfg), h.Reassign)
	}
}
