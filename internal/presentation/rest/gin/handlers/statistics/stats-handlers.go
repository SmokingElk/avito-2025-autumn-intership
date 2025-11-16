package statshandlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/SmokingElk/avito-2025-autumn-intership/docs"
	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/interfaces"
	"github.com/gin-gonic/gin"
)

type StatsHandlers struct {
	statsService interfaces.StatsService
}

func CreateStatsHandlers(statsService interfaces.StatsService) *StatsHandlers {
	return &StatsHandlers{
		statsService: statsService,
	}
}

// Add godoc
// @Summary Получить статистику назначений пользователей ревьюверами
// @Tags Stats
// @Param limit query int true "Количество записей в результате"
// @Param offset query int true "Отступ в статистике"
// @Produce json
// @Success 200 {object} docs.AssignmentsStats "Статистика по назначениям"
// @Router /stats/assignmentsPerMember [get]
func (h *StatsHandlers) GetAssignmentsPerMember(ctx *gin.Context) {
	limitStr := ctx.Query("limit")
	limit, err := strconv.Atoi(limitStr)

	if limitStr == "" || err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, docs.NewErrorResponse(
			"BAD_REQUEST",
			"invalid limit param",
		))
		return
	}

	offsetStr := ctx.Query("offset")
	offset, err := strconv.Atoi(offsetStr)

	if offsetStr == "" || err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, docs.NewErrorResponse(
			"BAD_REQUEST",
			"invalid offset param",
		))
		return
	}

	stats, err := h.statsService.GetAssignmentsPerMember(ctx.Request.Context(), limit, offset)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, docs.NewErrorResponse(
			"INTERNAL_SERVER_ERROR",
			fmt.Sprintf("failed to get assignmnets per member: %s", err.Error()),
		))

		return
	}

	resp := docs.AssignmentsStats{
		Count:   len(stats),
		Results: make([]docs.AssignmentsPerMember, 0, len(stats)),
	}

	for _, assignmentsStats := range stats {
		resp.Results = append(resp.Results, docs.ToAssignmentsPerMemberResponse(assignmentsStats))
	}

	ctx.JSON(http.StatusOK, resp)
}

func InitStatsHandlers(r *gin.RouterGroup, statsService interfaces.StatsService) {
	h := CreateStatsHandlers(statsService)

	group := r.Group("stats")

	{
		group.GET("assignmentsPerMember", h.GetAssignmentsPerMember)
	}
}
