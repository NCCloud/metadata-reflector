package common

import (
	"time"

	"github.com/caarlos0/env/v11"
)

//go:generate go run github.com/g4s8/envdoc@latest -output ../../environments.md -type Config
type Config struct {
	// the interval of the background propagation task
	BackgroundReflectionInterval time.Duration `env:"BACKGROUND_REFLECTION_INTERVAL" envDefault:"5m"`
	// a deployment selector to limit the watched resources
	// should be provided in this format https://pkg.go.dev/k8s.io/apimachinery/pkg/labels#Parse
	// if empty, all deployments will match
	DeploymentSelector string `env:"DEPLOYMENT_SELECTOR" envDefault:""`
	// a comma-separated list of namespaces where to watch the deployments
	// if empty, all namespaces will be watched
	Namespaces []string `env:"NAMESPACES" envDefault:""`
	// the port on which the Prometheus server should be exposed
	PrometheusMetricsPort int `env:"PROMETHEUS_METRICS_PORT" envDefault:"9090"`
	// the port for health checking
	HealthCheckPort int `env:"HEALTH_CHECK_PORT" envDefault:"8083"`
}

func NewConfig() *Config {
	config := &Config{}
	Must(env.Parse(config))

	return config
}
