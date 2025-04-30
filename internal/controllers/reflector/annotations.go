package reflector

import (
	"context"
	"strings"

	"github.com/NCCloud/metadata-reflector/internal/common"
	"github.com/hashicorp/go-multierror"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *Controller) reconcileAnnotations(ctx context.Context, deployment *appsv1.Deployment) (ctrl.Result, error) {
	r.logger.Info("Starting annotation reconciliation",
		"deployment", deployment.Name, "namespace", deployment.Namespace)
	defer r.logger.Info("Finished annotation reconciliation",
		"deployment", deployment.Name, "namespace", deployment.Namespace)

	var (
		annotationReflectResult ctrl.Result
		annotationReflectError  error
	)

	if common.MapHasPrefix(ReflectorAnnotationsAnnotationDomain, deployment.Annotations) {
		annotationReflectResult, annotationReflectError = r.reflectAnnotations(ctx, deployment)
	} else {
		annotationReflectResult, annotationReflectError = r.unsetReflectedAnnotations(ctx, deployment)
	}

	return annotationReflectResult, annotationReflectError
}

// reflect configuration from deployment to managed pods.
func (r *Controller) reflectAnnotations(ctx context.Context, deployment *appsv1.Deployment,
) (ctrl.Result, error) {
	deploymentName := deployment.Name

	// a map of reflector annotations present on the object
	reflectorAnnotations := common.FindPartialKeys(
		ReflectorAnnotationsAnnotationDomain, deployment.Annotations)

	annotationsToReflect, annotationsErr := r.keysToReflect(
		reflectorAnnotations, deployment.Annotations)
	if annotationsErr != nil {
		r.logger.Error(
			annotationsErr, "Could not get annotations to reflect",
			"deployment", deploymentName,
		)

		return ctrl.Result{}, annotationsErr
	}

	// nothing to reflect, let's try to unset reflected annotations
	if len(annotationsToReflect) == 0 {
		return r.unsetReflectedAnnotations(ctx, deployment)
	}

	specialReflectorAnn := r.getReflectorAnnForAnnotations(common.MapKeysAsString(annotationsToReflect))

	for key, value := range specialReflectorAnn {
		annotationsToReflect[key] = value
	}

	pods, podListError := r.getManagedPods(ctx, deployment)
	if podListError != nil {
		r.logger.Error(podListError,
			"Error listing pods for deployment",
			"deployment", deploymentName,
		)

		return ctrl.Result{}, podListError
	}

	var podUpdateErrors *multierror.Error

	for _, pod := range pods.Items {
		shouldUpdatePod := false

		if annotationsUpdated := r.setAnnotations(annotationsToReflect, &pod); annotationsUpdated {
			shouldUpdatePod = true
		}

		excessiveAnnotationsUnset := r.unsetExcessiveAnnotations(annotationsToReflect, &pod)
		if excessiveAnnotationsUnset {
			shouldUpdatePod = true
		}

		if !shouldUpdatePod {
			continue
		}

		updateErr := r.kubeClient.UpdatePod(ctx, pod)
		if updateErr != nil {
			r.logger.Error(updateErr, "Failed to update pod metadata",
				"pod", pod.Name,
			)

			podUpdateErrors = multierror.Append(podUpdateErrors, updateErr)
		}
	}

	return ctrl.Result{}, podUpdateErrors.ErrorOrNil()
}

// generate a map of annotations that help identify what annotations
// were set by the reflector on the managed object.
func (r *Controller) getReflectorAnnForAnnotations(labelsToReflect string) map[string]string {
	annotationsToReflect := make(map[string]string)
	annotationsToReflect[ReflectorAnnotationsReflectedAnnotation] = labelsToReflect

	return annotationsToReflect
}

// reflect annotation to managed pods.
// returns whether any pod annotation was updated.
func (r *Controller) setAnnotations(annotations map[string]string, pod *v1.Pod) bool {
	podUpdated := false

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	for key, value := range annotations {
		annotationValue, annotationOk := pod.Annotations[key]
		if annotationOk && annotationValue == value {
			continue
		}

		r.logger.Info("Setting annotation for pod", "pod", pod.Name, "annotation", key, "value", value)

		pod.Annotations[key] = value
		podUpdated = true
	}

	return podUpdated
}

// unset managed annotations from a pod.
// returns whether any annotation was unset.
func (r *Controller) unsetAnnotations(annotations []string, pod *v1.Pod) bool {
	anyAnnotationUnset := false

	if pod.Annotations == nil {
		return anyAnnotationUnset
	}

	for _, annotation := range annotations {
		if _, annotationExists := pod.Annotations[annotation]; !annotationExists {
			continue
		}

		r.logger.Info("Unsetting annotation from pod", "pod", pod.Name, "annotation", annotation)

		delete(pod.Annotations, annotation)

		anyAnnotationUnset = true
	}

	return anyAnnotationUnset
}

// unset excessive annotations given the annotations we expect to be present on an object.
// returns if any annotation was unset.
func (r *Controller) unsetExcessiveAnnotations(annotationsToReflect map[string]string, pod *v1.Pod) bool {
	var annotationsToUnset []string

	podReflectedAnnValue, podReflectedAnnExists := pod.Annotations[ReflectorAnnotationsReflectedAnnotation]
	if !podReflectedAnnExists {
		return false
	}

	currentKeys := strings.Split(podReflectedAnnValue, ",")
	expectedKeys := common.MapKeysAsSlice(annotationsToReflect)
	annotationsToUnset = common.ExcessiveElements(expectedKeys, currentKeys)

	if len(annotationsToUnset) == 0 {
		return false
	}

	if annotationsUpdated := r.unsetAnnotations(annotationsToUnset, pod); annotationsUpdated {
		return annotationsUpdated
	}

	return false
}

func (r *Controller) unsetReflectedAnnotations(ctx context.Context, deployment *appsv1.Deployment,
) (ctrl.Result, error) {
	deploymentName := deployment.Name

	pods, podListError := r.getManagedPods(ctx, deployment)
	if podListError != nil {
		r.logger.Error(podListError,
			"Error listing pods for deployment",
			"deployment", deploymentName,
		)

		return ctrl.Result{}, podListError
	}

	var podUpdateErrors *multierror.Error

	for _, pod := range pods.Items {
		annotationValue, hasReflectorAnnotation := pod.Annotations[ReflectorAnnotationsReflectedAnnotation]
		// if the annotation is not present, configuration is either already unset
		// or the annotation was deleted manually and we don't know what annotations to remove
		if !hasReflectorAnnotation {
			continue
		}

		shouldUpdatePod := false

		annotationsToUnset := strings.Split(annotationValue, ",")
		annotationsToUnset = append(annotationsToUnset, ReflectorAnnotationsReflectedAnnotation)

		if annotationsUpdated := r.unsetAnnotations(annotationsToUnset, &pod); annotationsUpdated {
			shouldUpdatePod = true
		}

		if !shouldUpdatePod {
			continue
		}

		updateErr := r.kubeClient.UpdatePod(ctx, pod)
		if updateErr != nil {
			r.logger.Error(updateErr, "Failed to unset metadata from pod",
				"pod", pod.Name,
			)

			podUpdateErrors = multierror.Append(podUpdateErrors, updateErr)
		}
	}

	return ctrl.Result{}, podUpdateErrors.ErrorOrNil()
}
