package teamhandlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/SmokingElk/avito-2025-autumn-intership/docs"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	teamErrors "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/errors"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/interfaces"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/auth"
	request_id "github.com/SmokingElk/avito-2025-autumn-intership/internal/presentation/rest/gin/middleware/request-id"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type TeamHandlers struct {
	teamService interfaces.TeamService
	logger      zerolog.Logger
}

func CreateTeamHandlers(teamService interfaces.TeamService, log zerolog.Logger) *TeamHandlers {
	return &TeamHandlers{
		teamService: teamService,
		logger:      log,
	}
}

// Add godoc
// @Summary Создать команду с участниками (создает/обновляет пользователей)
// @Tags Teams
// @Accept json
// @Produce json
// @Param input body docs.AddTeamRequest true "Данные для создания/обновления"
// @Success 201 {object} docs.AddTeamResponse "Команда создана"
// @Failure 400 {object} docs.ErrorResponse "Команда уже существует"
// @Failure 401 {object} docs.ErrorResponse "Нет/неверный админский токен"
// @Failure 409 {object} docs.ErrorResponse "Пользователь является членом другой команды"
// @Router /team/add [post]
func (h *TeamHandlers) Add(ctx *gin.Context) {
	log := h.localLogger(ctx, "Add")

	var request docs.AddTeamRequest

	if err := ctx.BindJSON(&request); err != nil {
		log.Warn().Msg("invalid body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, docs.NewErrorResponse(
			"BAD_REQUEST",
			"invalid body",
		))
		return
	}

	membersEntities := make([]memberEntity.Member, 0, len(request.Members))

	for _, member := range request.Members {
		membersEntities = append(membersEntities, member.ToTeamMemberEntity())
	}

	err := h.teamService.Upsert(ctx.Request.Context(), request.Name, membersEntities)

	if err != nil {
		switch {
		case errors.Is(err, teamErrors.ErrTeamExists):
			log.Warn().Msg("team already exists")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, docs.NewErrorResponse(
				"TEAM_EXISTS",
				"team_name already exists",
			))

		case errors.Is(err, teamErrors.ErrMemberOfOtherTeam):
			log.Warn().Msg("member of other team")
			ctx.AbortWithStatusJSON(http.StatusConflict, docs.NewErrorResponse(
				"MEMBER_OF_OTHER_TEAM",
				"User is member of other team",
			))

		default:
			log.Error().Err(err).Msg("failed to create team")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
				"INTERNAL_SERVER_ERROR",
				fmt.Sprintf("failed to create team: %s", err.Error()),
			))
		}

		return
	}

	resp := docs.AddTeamResponse{
		Team: docs.AddTeamResponseObject(request),
	}

	ctx.JSON(http.StatusCreated, resp)

	log.Info().Msg("successfully added team")
}

// Add godoc
// @Summary Получить команду с участниками
// @Tags Teams
// @Security BearerAuth
// @Param team_name query string true "Уникальное имя команды"
// @Produce json
// @Success 200 {object} docs.GetTeamResponse "Объект команды"
// @Failure 401 {object} docs.ErrorResponse "Нет/неверный админский токен"
// @Failure 404 {object} docs.ErrorResponse "Команда не найдена"
// @Router /team/get [get]
func (h *TeamHandlers) Get(ctx *gin.Context) {
	log := h.localLogger(ctx, "Get")

	teamName := ctx.Query("team_name")

	if teamName == "" {
		log.Warn().Msg("invalid team_name param")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, docs.NewErrorResponse(
			"BAD_REQUEST",
			"invalid team_name param",
		))
		return
	}

	team, err := h.teamService.GetByName(ctx.Request.Context(), teamName)

	if err != nil {
		switch {
		case errors.Is(err, teamErrors.ErrTeamNotFound):
			log.Warn().Msg("team not found")
			ctx.AbortWithStatusJSON(http.StatusNotFound, docs.NewErrorResponse(
				"NOT_FOUND",
				"resource not found",
			))

		default:
			log.Error().Err(err).Msg("failed to get team")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
				"INTERNAL_SERVER_ERROR",
				fmt.Sprintf("failed to get team: %s", err.Error()),
			))
		}

		return
	}

	resp := docs.GetTeamResponse{
		Name:    team.Name,
		Members: make([]docs.TeamMember, 0, len(team.Members)),
	}

	for _, member := range team.Members {
		resp.Members = append(resp.Members, docs.ToTeamMember(member))
	}

	ctx.JSON(http.StatusOK, resp)

	log.Info().Msg("successfully created team")
}

func (h *TeamHandlers) localLogger(ctx *gin.Context, opName string) zerolog.Logger {
	log := h.logger.With().
		Str("op", opName).
		Str("requestId", ctx.GetString(request_id.REQUEST_ID_PARAM)).
		Logger()

	return log
}

func InitTeamHandlers(r *gin.RouterGroup, log zerolog.Logger, teamService interfaces.TeamService, cfg *config.RestConfig) {
	h := CreateTeamHandlers(teamService, log)

	group := r.Group("team")

	{
		group.POST("add", h.Add)
		group.GET("get", auth.WithAuth(cfg), h.Get)
	}
}
