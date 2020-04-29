package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber/jaeger-lib/metrics"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
)

const (
	defaultMetricsRoute = "/metrics"
)

// New ...
func New(metricsRoute string) *Builder {
	if metricsRoute == "" {
		metricsRoute = defaultMetricsRoute
	}
	return &Builder{HTTPRoute: metricsRoute}
}

// Builder provides command line options to configure metrics backend used by Jaeger executables.
type Builder struct {
	HTTPRoute string // endpoint name to expose metrics, e.g. for scraping
	handler   http.Handler
}

// CreateMetricsFactory creates a metrics factory based on the configured type of the backend.
// If the metrics backend supports HTTP endpoint for scraping, it is stored in the builder and
// can be later added by RegisterHandler function.
func (b *Builder) CreateMetricsFactory(namespace string) (metrics.Factory, error) {
	metricsFactory := jprom.New().Namespace(metrics.NSOptions{Name: namespace, Tags: nil})
	b.handler = promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{DisableCompression: true})
	return metricsFactory, nil
}

// Handler returns an http.Handler for the metrics endpoint.
func (b *Builder) Handler() http.Handler {
	return b.handler
}
