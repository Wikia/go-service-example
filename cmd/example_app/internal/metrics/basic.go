package metrics

import "github.com/prometheus/client_golang/prometheus"

var GreetCount = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "greets_total",
		Help: "Number of generated greetings",
	})


func RegisterMetrics(registerer prometheus.Registerer) {
	registerer.MustRegister(GreetCount)
}