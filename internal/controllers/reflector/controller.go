package reflector

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/NCCloud/metadata-reflector/internal/clients"
	"github.com/NCCloud/metadata-reflector/internal/common"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "k8s.io/api/core/v1"
)

type ReflectorController struct {
	kubeClient clients.KubernetesClient
	logger     logr.Logger
	config     *common.Config
}

func NewReflectorController(
	kubeClient clients.KubernetesClient, logger logr.Logger, config *common.Config,
) ReflectorController {
	return ReflectorController{
		kubeClient: kubeClient,
		logger:     logger,
		config:     config,
	}
}

func (r *ReflectorController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	namespacedName := req.NamespacedName
	r.logger.Info("Starting reconciliation", "namespacedName", namespacedName)
	defer r.logger.Info("Finished reconciliation", "namespacedName", namespacedName)

	deployment, getDeployErr := r.kubeClient.GetDeployment(ctx, namespacedName)
	if getDeployErr != nil {
		r.logger.Error(getDeployErr, "Failed to get deployment", "namespacedName", namespacedName)
		return ctrl.Result{}, getDeployErr
	}

	var reflectorErrors []error

	var (
		labelReflectResult ctrl.Result
		labelReflectError  error
	)
	if common.MapHasPrefix(ReflectorLabelsAnnotationDomain, deployment.Annotations) {
		labelReflectResult, labelReflectError = r.ReflectLabels(ctx, deployment)
	} else {
		labelReflectResult, labelReflectError = r.UnsetReflectedLabels(ctx, deployment)
	}

	// if the error is not nil, it always takes precedence over the result
	// the idea is to not requeue after any error as there can be some errors that we want to skip
	// but also maintain the possibility to requeue now/after some time when there was no error
	// and the result explicitly states that we need to requeue
	if labelReflectError != nil {
		reflectorErrors = append(reflectorErrors, labelReflectError)
	} else if r.shouldRequeueNow(labelReflectResult) {
		return labelReflectResult, labelReflectError
	}

	// check if there is an error that requires requeuing
	// we want to skip some errors that cannot be fixed by the controller
	for _, err := range reflectorErrors {
		if r.shouldRequeueAfterError(err) {
			r.logger.Info("Reconciliation didn't succeed, requeuing...",
				"namespacedName", namespacedName)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// check if the error can be skipped as it's not fixable
func (r *ReflectorController) shouldRequeueAfterError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrUnparsableAnnotation) {
		return false
	}
	// since deploymentSelector is an immutable field, it will not be updated,
	// so we will never be able to get pods
	if errors.Is(err, ErrEmptyPodSelector) {
		return false
	}
	return true
}

func (r *ReflectorController) shouldRequeueNow(result ctrl.Result) bool {
	if result.Requeue || result.RequeueAfter != 0 {
		return true
	}

	return false
}

// get a list of pods managed by deployment
func (r *ReflectorController) GetManagedPods(
	ctx context.Context, deployment *appsv1.Deployment,
) (*v1.PodList, error) {
	deploymentName := deployment.Name
	podSelector, selectorErr := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if selectorErr != nil {
		r.logger.Error(selectorErr,
			"Failed to get managed pods", "deployment", deploymentName)
		return nil, selectorErr
	}

	if podSelector.Empty() {
		r.logger.Error(ErrEmptyPodSelector,
			"Cannot get managed pods as the selector would match everything",
			"deployment", deploymentName)
		return nil, ErrEmptyPodSelector
	}

	pods, podListError := r.kubeClient.ListPods(ctx, podSelector)
	if podListError != nil {
		return nil, podListError
	}

	if len(pods.Items) == 0 {
		podNotFoundErr := fmt.Errorf(
			"Could not find pods for deployment %s with selector %s",
			deploymentName, podSelector.String())
		return nil, podNotFoundErr
	}
	r.logger.Info("Found Managed pods",
		"count", len(pods.Items), "selector", podSelector.String())
	return pods, nil
}

func (r *ReflectorController) FilterCreateEvents(e event.CreateEvent) bool {
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

func (r *ReflectorController) FilterUpdateEvents(e event.UpdateEvent) bool {
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

func (r *ReflectorController) SetupWithManager(mgr ctrl.Manager) error {

	predicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.FilterCreateEvents(e)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.FilterUpdateEvents(e)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(predicate).
		Complete(r)
}
