package auth

import (
	"net/http"
	"strings"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/config"
	"github.com/gin-gonic/gin"
)

const bearerPrefix = "Bearer "

func WithAuth(cfg *config.RestConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.Request.Header.Get("Authorization")

		if !strings.HasPrefix(tokenStr, bearerPrefix) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(tokenStr, bearerPrefix)

		// mock auth service integration
		authorized := token == cfg.AdminToken

		if !authorized {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Next()
	}
}
