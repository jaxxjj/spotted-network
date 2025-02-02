package metric

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Basic metrics
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spotted_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method", "endpoint"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "spotted_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spotted_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type"},
	)
)

type Server struct {
	conf *Config
}

type Config struct {
	Port int `default:"4014"`
}

func New(conf *Config) *Server {
	if conf == nil {
		conf = &Config{}
		envconfig.MustProcess("metric", conf)
	}
	return &Server{conf: conf}
}

func (s *Server) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.conf.Port), nil)
}

// RecordRequest records a request metric
func RecordRequest(method, endpoint string) {
	requestsTotal.WithLabelValues(method, endpoint).Inc()
}

// RecordRequestDuration records the duration of a request
func RecordRequestDuration(method, endpoint string, duration time.Duration) {
	requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordError records an error metric
func RecordError(errorType string) {
	errorsTotal.WithLabelValues(errorType).Inc()
}
