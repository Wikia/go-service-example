package tracing

import (
	"io"

	"github.com/Wikia/go-example-service/internal/logging"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	jaeger_config "github.com/uber/jaeger-client-go/config"
	jaeger_metrics "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
)

func InitJaegerTracer(serviceName string, logger *zap.SugaredLogger, registry prometheus.Registerer) (tracer opentracing.Tracer, closer io.Closer, err error) {
	traceCfg, err := jaeger_config.FromEnv()

	if err != nil {
		return nil, nil, errors.Wrap(err, "could not initialize tracer configuration")
	}

	traceCfg.ServiceName = serviceName
	tracingLogger := &logging.TracingLogger{Logger: logger}
	metricsFactory := jaeger_metrics.New(jaeger_metrics.WithRegisterer(registry))
	tracer, closer, err = traceCfg.NewTracer(
		jaeger_config.Logger(tracingLogger),
		jaeger_config.Metrics(metricsFactory),
	)

	if err == nil {
		opentracing.SetGlobalTracer(tracer)
	}

	return
}
