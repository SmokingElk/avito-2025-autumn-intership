package request_id

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const REQUEST_ID_PARAM = "__request_id_param"

func AddRequestId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestId := uuid.New().String()

		ctx.Set(REQUEST_ID_PARAM, requestId)

		ctx.Next()
	}
}
