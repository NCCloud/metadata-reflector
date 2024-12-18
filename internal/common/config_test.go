package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig_SuccessfulParse(t *testing.T) {
	deploymentSelector := "app=test"

	os.Setenv("DEPLOYMENT_SELECTOR", deploymentSelector)
	defer os.Unsetenv("DEPLOYMENT_SELECTOR")

	config := NewConfig()

	assert.Equal(t, config.DeploymentSelector, deploymentSelector)
}

func TestNewConfig_UnsuccessfulParse(t *testing.T) {
	invalidPrometheusPort := "nine-thousand-ninety"

	os.Setenv("PROMETHEUS_METRICS_PORT", invalidPrometheusPort)
	defer os.Unsetenv("PROMETHEUS_METRICS_PORT")

	assert.Panics(t, func() { NewConfig() })
}
