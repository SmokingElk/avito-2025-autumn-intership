package healthhandlers

import (
	"net/http"

	"github.com/SmokingElk/avito-2025-autumn-intership/docs"
	"github.com/gin-gonic/gin"
)

type HealthHandlers struct {
}

func CreateHealthHandlers() *HealthHandlers {
	return &HealthHandlers{}
}

// Add godoc
// @Summary Проверка работоспособности сервиса
// @Tags Health
// @Produce json
// @Success 200 {object} docs.HealthResponse "Service healthy"
// @Router /health [get]
func (h *HealthHandlers) Health(ctx *gin.Context) {
	resp := docs.HealthResponse{
		Name:   "pull-request service",
		Status: "ok",
	}

	ctx.JSON(http.StatusOK, resp)
}

func InitHealthHandlers(r *gin.RouterGroup) {
	h := CreateHealthHandlers()

	r.GET("/health", h.Health)
}
