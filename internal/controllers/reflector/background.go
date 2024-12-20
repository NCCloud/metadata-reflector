package reflector

import (
	"context"

	"github.com/NCCloud/metadata-reflector/internal/metrics"

	"github.com/go-co-op/gocron/v2"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// periodically check if reflected metadata is present on pods.
func (r *Controller) StartBackgroundJob(ctx context.Context) error {
	r.logger.Info("Starting the reflector gocron job",
		"secondsInterval", r.config.BackgroundReflectionInterval)

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		r.logger.Error(err, "Failed to create the reflector scheduler")

		return err
	}

	defer func() { _ = scheduler.Shutdown() }()

	_, err = scheduler.NewJob(
		gocron.DurationJob(r.config.BackgroundReflectionInterval),
		gocron.NewTask(
			func() {
				if err := r.reflectBackground(ctx); err != nil {
					r.logger.Error(err,
						"Background reflector task failed to complete")
				}
			},
		),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		r.logger.Error(err, "Failed to create the reflector scheduler job")

		return err
	}

	scheduler.Start()

	select {}
}

func (r *Controller) reflectBackground(ctx context.Context) error {
	r.logger.Info("Starting reflector background task")

	deploymentSelector, parseErr := labels.Parse(r.config.DeploymentSelector)
	if parseErr != nil {
		r.logger.Error(parseErr, "Failed to parse deployment selector")

		return parseErr
	}

	deployments, depListErr := r.kubeClient.ListDeployments(ctx, deploymentSelector)
	if depListErr != nil {
		r.logger.Error(depListErr,
			"Failed to list deployments",
			"selector", deploymentSelector.String(),
			"error", depListErr,
		)

		return depListErr
	}

	deploymentCount := len(deployments.Items)
	r.logger.Info(
		"Extracted deployments",
		"count", deploymentCount,
		"selector", deploymentSelector.String(),
	)

	metrics.DeploymentsMatchingSelector.WithLabelValues(
		deploymentSelector.String()).Set(float64(deploymentCount))

	for _, deployment := range deployments.Items {
		namespacedName := types.NamespacedName{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
		}
		request := ctrl.Request{NamespacedName: namespacedName}

		_, reconcileErr := r.Reconcile(ctx, request)
		if reconcileErr != nil {
			r.logger.Error(reconcileErr,
				"Background reconciliation failed", "namespacedName", namespacedName)
		}
	}

	r.logger.Info("Finished reflector background task")

	return nil
}
