package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Readiness(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}
