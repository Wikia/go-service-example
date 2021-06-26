package handlers

import (
	"github.com/Wikia/go-example-service/internal/logging"
	"net/http"

	"github.com/Wikia/go-example-service/cmd/server/metrics"
	"github.com/labstack/echo/v4"
)

type Message struct {
	Text string
}

func Hello(ctx echo.Context) error {
	logger := logging.FromEchoContext(ctx)
	logger.Info("Greeting user")
	defer metrics.GreetCount.Inc()

	m := Message{"Hello World"}
	return ctx.JSON(http.StatusOK, m)
}
