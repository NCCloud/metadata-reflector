package main

import (
	"fmt"
	"os"

	"github.com/NCCloud/metadata-reflector/internal/common"
	"github.com/NCCloud/metadata-reflector/internal/controllers/reflector"

	"github.com/NCCloud/metadata-reflector/internal/clients"

	ctrl "sigs.k8s.io/controller-runtime"

	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	config := common.NewConfig()

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.LogLevel)); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level %q, defaulting to info, error: %v\n", config.LogLevel, err)
		// default to info if invalid
		level = zapcore.InfoLevel
	}

	opts := zap.Options{
		Development: false,
		Level:       level,
	}

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	logger.Info("Log level configuration", "configured", level.String())

	mgr, mgrErr := clients.NewControllerManager(config, logger)
	if mgrErr != nil {
		logger.Error(mgrErr, "Failed to created the controller manager")
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

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "problem running manager")
		panic(err)
	}
}
