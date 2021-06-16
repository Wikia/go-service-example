package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}
