package reflector

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/NCCloud/metadata-reflector/internal/common"
	"github.com/hashicorp/go-multierror"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// reflect configuration from deployment to managed pods.
func (r *Controller) reflectLabels(ctx context.Context, deployment *appsv1.Deployment,
) (ctrl.Result, error) {
	deploymentName := deployment.Name

	// a map of reflector annotations present on the object
	reflectorAnnotations := common.FindPartialKeys(
		ReflectorLabelsAnnotationDomain, deployment.Annotations)

	labelsToReflect, labelsErr := r.labelsToReflect(
		reflectorAnnotations, deployment.Labels)
	if labelsErr != nil {
		r.logger.Error(labelsErr, "Could not get labels to reflect", "deployment", deploymentName)

		return ctrl.Result{}, labelsErr
	}

	// nothing to reflect, let's try to unset reflected labels
	if len(labelsToReflect) == 0 {
		return r.unsetReflectedLabels(ctx, deployment)
	}

	reflectedAnnotations := r.getReflectedAnnotations(common.MapKeysAsString(labelsToReflect))

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

		if labelsUpdated := r.setLabels(labelsToReflect, &pod); labelsUpdated {
			shouldUpdatePod = true
		}

		if excessiveLabelsUnset := r.unsetExcessiveLabels(labelsToReflect, &pod); excessiveLabelsUnset {
			shouldUpdatePod = true
		}

		if annotationsUpdated := r.setAnnotations(reflectedAnnotations, &pod); annotationsUpdated {
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

func (r *Controller) unsetReflectedLabels(ctx context.Context, deployment *appsv1.Deployment,
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
		annotationValue, hasReflectorAnnotation := pod.Annotations[ReflectorLabelsReflectedAnnotation]
		// if the annotation is not present, configuration is either already unset
		// or the annotation was deleted manually and we don't know what labels to remove
		if !hasReflectorAnnotation {
			continue
		}

		shouldUpdatePod := false

		labelsToUnset := strings.Split(annotationValue, ",")
		if labelsUpdated := r.unsetLabels(labelsToUnset, &pod); labelsUpdated {
			shouldUpdatePod = true
		}

		annotationsToUnset := []string{ReflectorLabelsReflectedAnnotation}
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

// reflect a list of labels from deployment to the pod.
// return whether any pod label was updated.
func (r *Controller) setLabels(labels map[string]string, pod *v1.Pod) bool {
	podUpdated := false

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}

	for key, value := range labels {
		podLabelValue, ok := pod.Labels[key]
		if ok && value == podLabelValue {
			continue
		}

		r.logger.Info("Setting pod label", "pod",
			pod.Name, "label", key, "value", value)

		pod.Labels[key] = value
		podUpdated = true
	}

	return podUpdated
}

// unset managed labels from a pod.
// returns whether any label was unset.
func (r *Controller) unsetLabels(labels []string, pod *v1.Pod) bool {
	anyLabelDeleted := false

	for _, label := range labels {
		if _, ok := pod.Labels[label]; !ok {
			continue
		}

		r.logger.Info("Unsetting label from pod", "pod", pod.Name, "label", label)

		delete(pod.Labels, label)

		anyLabelDeleted = true
	}

	return anyLabelDeleted
}

// given reflector annotations from the object and a map of labels, find labels that need to be reflected.
func (r *Controller) labelsToReflect(
	reflectorAnnotations map[string]string, labels map[string]string,
) (map[string]string, error) {
	expectedAnnotationParts := 2
	labelsToReflect := make(map[string]string)

	for annKey, annValue := range reflectorAnnotations {
		annotationKeyParts := strings.Split(annKey, "/")
		if len(annotationKeyParts) != expectedAnnotationParts {
			r.logger.Error(ErrUnparsableAnnotation,
				"Annotation should consist of exactly 2 parts",
				"annotation", annKey, "parts", len(annotationKeyParts),
			)

			return nil, ErrUnparsableAnnotation
		}

		switch operation := annotationKeyParts[1]; operation {
		case ReflectorLabelsList:
			annotationLabels := strings.Split(annValue, ",")

			for key, value := range labels {
				if !slices.Contains(annotationLabels, key) {
					continue
				}

				labelsToReflect[key] = value
			}

		case ReflectorLabelsRegex:
			pattern := common.ExactMatchRegex(annValue)
			regex := regexp.MustCompile(pattern)

			for key, value := range labels {
				if !regex.MatchString(key) {
					continue
				}

				labelsToReflect[key] = value
			}

		default:
			r.logger.Error(ErrUnparsableAnnotation,
				"Annotation doesn't have a valid operation to parse", "annotation", annKey)

			return nil, ErrUnparsableAnnotation
		}
	}

	return labelsToReflect, nil
}

// unset excessive labels given the labels we expect to be present on an object.
// returns if any label was unset.
func (r *Controller) unsetExcessiveLabels(labelsToReflect map[string]string, pod *v1.Pod) bool {
	var labelsToUnset []string

	podReflectedAnnValue, podReflectedAnnExists := pod.Annotations[ReflectorLabelsReflectedAnnotation]
	if !podReflectedAnnExists {
		return false
	}

	currentKeys := strings.Split(podReflectedAnnValue, ",")
	expectedKeys := common.MapKeysAsSlice(labelsToReflect)
	labelsToUnset = common.ExcessiveElements(expectedKeys, currentKeys)

	if len(labelsToUnset) == 0 {
		return false
	}

	if labelsUpdated := r.unsetLabels(labelsToUnset, pod); labelsUpdated {
		return labelsUpdated
	}

	return false
}
