package handlers

import (
	"net/http"

	"github.com/Wikia/go-example-service/cmd/server/metrics"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Message struct {
	Text string
}

func Hello(ctx echo.Context) error {
	logger := zap.S()
	logger.Info("Greeting user")
	defer metrics.GreetCount.Inc()

	m := Message{"Hello World"}
	return ctx.JSON(http.StatusOK, m)
}
