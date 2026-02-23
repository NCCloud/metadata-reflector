package clients

import (
	"fmt"

	"github.com/NCCloud/metadata-reflector/internal/common"
	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerConfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func NewControllerManager(config *common.Config, logger logr.Logger) (manager.Manager, error) {
	scheme := runtime.NewScheme()

	common.Must(clientgoscheme.AddToScheme(scheme))

	cacheOptions, cacheOptsErr := GetCacheOptions(config, logger)
	if cacheOptsErr != nil {
		return nil, errors.Wrap(cacheOptsErr, "failed to get cache options")
	}

	mgr, managerErr := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Logger: logger,
		Metrics: server.Options{
			BindAddress: fmt.Sprintf(":%d", config.PrometheusMetricsPort),
		},
		HealthProbeBindAddress: fmt.Sprintf(":%d", config.HealthCheckPort),
		LeaderElection:         config.EnableLeaderElection,
		LeaderElectionID:       "metadata-reflector-leader.spaceship.com",
		Cache:                  cacheOptions,
		Controller: controllerConfig.Controller{
			MaxConcurrentReconciles: config.MaxConcurrentReconciles,
		},
	})
	if managerErr != nil {
		return nil, errors.Wrap(managerErr, "failed to get manager")
	}

	return mgr, nil
}

func GetCacheOptions(config *common.Config, logger logr.Logger) (cache.Options, error) {
	var labelParseErr error

	rawDeploymentSelector := config.DeploymentSelector
	labelSelector := labels.NewSelector()

	// if the selector is empty, it will only match resources without labels
	if rawDeploymentSelector != "" {
		labelSelector, labelParseErr = labels.Parse(rawDeploymentSelector)
		if labelParseErr != nil {
			logger.Error(labelParseErr,
				"Failed to construct deployment selector",
				"selector", config.DeploymentSelector,
			)

			return cache.Options{}, labelParseErr
		}

		logger.Info("DEPLOYMENT_SELECTOR is set, will only watch deployments matching it",
			"selector", labelSelector.String())
	}

	namespaces := make(map[string]cache.Config)

	if len(config.Namespaces) > 0 {
		for _, namespace := range config.Namespaces {
			namespaces[namespace] = cache.Config{}
		}

		logger.Info("NAMESPACES is set, will only watch resources in them",
			"namespaces", config.Namespaces)
	}

	return cache.Options{
		ByObject: map[client.Object]cache.ByObject{
			&appsv1.Deployment{}: {
				Label: labelSelector,
			},
		},
		DefaultNamespaces: namespaces,
	}, nil
}
