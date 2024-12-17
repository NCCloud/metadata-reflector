package main

import (
	"context"

	"github.com/NCCloud/metadata-reflector/internal/common"
	"github.com/NCCloud/metadata-reflector/internal/controllers/reflector"
	_ "github.com/NCCloud/metadata-reflector/internal/metrics"

	"github.com/NCCloud/metadata-reflector/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	ctx := context.Background()
	config := common.NewConfig()
	logger := zap.New()

	mgr, mgrErr := clients.NewControllerManager(config, logger)
	if mgrErr != nil {
		panic(mgrErr)
	}

	kubeClient := clients.NewKubernetesClient(mgr, config)
	reflectorController := reflector.NewController(kubeClient, logger, config)

	if reflectorControllerErr := reflectorController.SetupWithManager(mgr); reflectorControllerErr != nil {
		panic(reflectorControllerErr)
	}

	if addHealthCheckErr := mgr.AddHealthzCheck("healthz", healthz.Ping); addHealthCheckErr != nil {
		panic(addHealthCheckErr)
	}

	if addReadyCheckErr := mgr.AddReadyzCheck("readyz", healthz.Ping); addReadyCheckErr != nil {
		panic(addReadyCheckErr)
	}

	if config.BackgroundReflectionInterval != 0 {
		go func() {
			if err := reflectorController.StartBackgroundJob(ctx); err != nil {
				panic(err)
			}
		}()
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "problem running manager")
		panic(err)
	}
}
