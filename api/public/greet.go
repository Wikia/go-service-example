package public

import (
	"net/http"

	"github.com/Wikia/go-commons/logging"
	"github.com/Wikia/go-example-service/metrics"
	"github.com/labstack/echo/v4"
)

type Message struct {
	Text string
}

func (s APIServer) Greet(ctx echo.Context) error {
	logger := logging.FromEchoContext(ctx)
	logger.Info("Greeting user")

	defer metrics.GreetCount.Inc()

	m := Message{"Hello World!"}

	return ctx.JSON(http.StatusOK, m)
}
