package clients

import (
	"fmt"
	"testing"

	"github.com/NCCloud/metadata-reflector/internal/common"
	mockManager "github.com/NCCloud/metadata-reflector/mocks/sigs.k8s.io/controller-runtime/pkg/manager"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	realManager "sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestNewControllerManager(t *testing.T) {
	logger := zap.New()

	// Mock the Kubernetes configuration retrieval
	mockConfig := &rest.Config{}
	originalGetConfigOrDie := ctrl.GetConfigOrDie
	ctrl.GetConfigOrDie = func() *rest.Config {
		return mockConfig
	}
	defer func() {
		// Restore the original GetConfigOrDie after the test
		ctrl.GetConfigOrDie = originalGetConfigOrDie
	}()

	// Mocked config for the test
	config := &common.Config{
		PrometheusMetricsPort: 8080,
		HealthCheckPort:       8081,
		EnableLeaderElection:  false,
		DeploymentSelector:    "app=test",
		Namespaces:            []string{"default", "test-namespace"},
	}

	// Mock the manager
	mockMgr := new(mockManager.MockManager)
	originalNewManager := ctrl.NewManager
	ctrl.NewManager = func(cfg *rest.Config, options ctrl.Options) (realManager.Manager, error) {
		assert.Equal(t, mockConfig, cfg) // Ensure the config is passed correctly
		assert.Equal(t, fmt.Sprintf(":%d", config.PrometheusMetricsPort), options.Metrics.BindAddress)
		assert.Equal(t, fmt.Sprintf(":%d", config.HealthCheckPort), options.HealthProbeBindAddress)
		assert.Equal(t, config.EnableLeaderElection, options.LeaderElection)
		assert.Equal(t, "metadata-reflector-leader.spaceship.com", options.LeaderElectionID)
		return mockMgr, nil
	}
	defer func() {
		// Restore the original NewManager after the test
		ctrl.NewManager = originalNewManager
	}()

	mgr, _ := NewControllerManager(config, logger)

	assert.NotNil(t, mgr)
	assert.Equal(t, mockMgr, mgr) // Ensure the returned manager matches the mock

}

func TestGetCacheOptions_SelectorAndNamespacesConfigured(t *testing.T) {
	logger := zap.New()

	rawSelector := "app=test"

	config := &common.Config{
		DeploymentSelector: rawSelector,
		Namespaces:         []string{"default", "test-namespace"},
	}

	options, _ := GetCacheOptions(config, logger)

	assert.NotNil(t, options)
	assert.Contains(t, options.ByObject, &appsv1.Deployment{})

	var byObjectDeployment cache.ByObject

	for key, value := range options.ByObject {
		if _, ok := key.(*appsv1.Deployment); ok {
			byObjectDeployment = value
		}
	}

	assert.Equal(t, rawSelector, byObjectDeployment.Label.String())
	assert.Contains(t, options.DefaultNamespaces, "default")
	assert.Contains(t, options.DefaultNamespaces, "test-namespace")

}

func TestGetCacheOptions_EmptyConfiguration(t *testing.T) {
	logger := zap.New()

	config := &common.Config{
		DeploymentSelector: "",
		Namespaces:         []string{},
	}

	options, _ := GetCacheOptions(config, logger)

	assert.NotNil(t, options)
	assert.Contains(t, options.ByObject, &appsv1.Deployment{})

	var byObjectDeployment cache.ByObject

	for key, value := range options.ByObject {
		if _, ok := key.(*appsv1.Deployment); ok {
			byObjectDeployment = value
		}
	}

	assert.Empty(t, byObjectDeployment.Label.String())
	assert.Empty(t, options.DefaultNamespaces)

}

func TestGetCacheOptions_InvalidConfiguration(t *testing.T) {
	logger := zap.New()

	config := &common.Config{
		DeploymentSelector: "invalid;selector",
		Namespaces:         []string{},
	}

	_, cacheOptsErr := GetCacheOptions(config, logger)

	assert.NotNil(t, cacheOptsErr)
}
