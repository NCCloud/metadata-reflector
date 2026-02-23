package reflector

import (
	"context"
	"reflect"

	"github.com/NCCloud/metadata-reflector/internal/clients"
	"github.com/NCCloud/metadata-reflector/internal/common"
	"github.com/hashicorp/go-multierror"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type Controller struct {
	kubeClient clients.KubernetesClient
	logger     logr.Logger
	config     *common.Config
}

func NewController(
	kubeClient clients.KubernetesClient, logger logr.Logger, config *common.Config,
) Controller {
	return Controller{
		kubeClient: kubeClient,
		logger:     logger,
		config:     config,
	}
}

func (r *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	namespacedName := req.NamespacedName

	r.logger.V(1).Info("Starting reconciliation", "namespacedName", namespacedName)
	defer r.logger.V(1).Info("Finished reconciliation", "namespacedName", namespacedName)

	deployment, getDeployErr := r.kubeClient.GetDeployment(ctx, namespacedName)
	if getDeployErr != nil {
		r.logger.Error(getDeployErr, "Failed to get deployment", "namespacedName", namespacedName)

		return ctrl.Result{}, getDeployErr
	}

	var reflectorErrors *multierror.Error

	labelReflectResult, labelReflectError := r.reconcileLabels(ctx, deployment)

	annReflectResult, annReflectError := r.reconcileAnnotations(ctx, deployment)

	reflectorErrors = multierror.Append(reflectorErrors, labelReflectError, annReflectError)

	// if the error is not nil, it always takes precedence over the result
	// the idea is not to requeue after any error as there can be other independent phases
	// but also maintain the possibility to requeue now/after some time when there was no error
	// and the result explicitly states that we need to requeue
	if r.shouldRequeueNow(labelReflectResult) {
		return labelReflectResult, labelReflectError
	} else if r.shouldRequeueNow(annReflectResult) {
		return annReflectResult, annReflectError
	}

	return ctrl.Result{RequeueAfter: r.config.BackgroundReflectionInterval}, reflectorErrors.ErrorOrNil()
}

func (r *Controller) FilterCreateEvents(e event.CreateEvent) bool {
	deployment, ok := e.Object.(*appsv1.Deployment)
	if !ok {
		return false
	}

	// check if the deployment contains any reflector annotation
	if common.MapContainsPartialKey(ReflectorAnnotationDomain, deployment.Annotations) {
		return true
	}

	return false
}

func (r *Controller) FilterUpdateEvents(e event.UpdateEvent) bool {
	newDeployment, ok := e.ObjectNew.(*appsv1.Deployment)
	if !ok {
		return false
	}

	oldDeployment, ok := e.ObjectOld.(*appsv1.Deployment)
	if !ok {
		return false
	}

	oldDepHasReflectorAnn := common.MapContainsPartialKey(ReflectorAnnotationDomain, oldDeployment.Annotations)
	newDepHasReflectorAnn := common.MapContainsPartialKey(ReflectorAnnotationDomain, newDeployment.Annotations)

	// the deployment doesn't have the reflector annotation
	if !oldDepHasReflectorAnn && !newDepHasReflectorAnn {
		return false
	}

	// annotations updated on deployment
	if !reflect.DeepEqual(newDeployment.Annotations, oldDeployment.Annotations) {
		return true
	}

	// deployment scaled, we need to re-apply labels
	newDeploymentReadyReplicas := newDeployment.Status.ReadyReplicas
	oldDeploymentReadyReplicas := oldDeployment.Status.ReadyReplicas

	if newDeploymentReadyReplicas > oldDeploymentReadyReplicas {
		return true
	}

	// labels updated on deployment
	if !reflect.DeepEqual(newDeployment.Labels, oldDeployment.Labels) {
		return true
	}

	return false
}

func (r *Controller) SetupWithManager(mgr ctrl.Manager) error {
	predicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.FilterCreateEvents(e)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.FilterUpdateEvents(e)
		},
		DeleteFunc: func(_ event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(_ event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(predicate).
		Complete(r)
}

func (r *Controller) shouldRequeueNow(result ctrl.Result) bool {
	return result.RequeueAfter != 0
}
