package ginlogger

import (
	"slices"
	"strings"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	"github.com/gin-gonic/gin"
)

func SkipLogger(cfg *config.RestConfig) gin.HandlerFunc {
	skipPaths := strings.Split(cfg.SkipLogging, ",")

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Skip: func(ctx *gin.Context) bool {
			return slices.Contains(skipPaths, ctx.Request.URL.Path)
		},
	})
}
