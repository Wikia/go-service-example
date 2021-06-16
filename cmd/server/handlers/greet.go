package handlers

import (
	"net/http"

	"github.com/Wikia/go-example-service/cmd/server/metrics"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Message struct {
	Text string
}

func Hello(ctx *gin.Context) {
	logger := zap.S()
	logger.Info("Greeting user")
	defer metrics.GreetCount.Inc()

	m := Message{"Hello World"}
	ctx.JSON(http.StatusOK, m)
}
