package reflector

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/NCCloud/metadata-reflector/internal/common"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// reflect configuration from deployment to managed pods
func (r *ReflectorController) ReflectLabels(ctx context.Context, deployment *appsv1.Deployment,
) (ctrl.Result, error) {
	namespacedName := fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)
	deploymentName := deployment.Name

	// a map of reflector annotations present on the object
	reflectorAnnotations := common.FindPartialKeys(
		ReflectorLabelsAnnotationDomain, deployment.Annotations)

	labelsToReflect, labelsErr := r.LabelsToReflect(
		reflectorAnnotations, deployment.Labels)
	if labelsErr != nil {
		r.logger.Error(labelsErr, "Could not get labels to reflect", "deployment", deploymentName)
		return ctrl.Result{}, labelsErr
	}

	// nothing to reflect, let's try to unset reflected labels
	if len(labelsToReflect) == 0 {
		return r.UnsetReflectedLabels(ctx, deployment)
	}

	reflectedAnnotations := r.GetReflectedAnnotations(common.MapKeysAsString(labelsToReflect))

	pods, podListError := r.GetManagedPods(ctx, deployment)
	if podListError != nil {
		r.logger.Error(podListError,
			"Error listing pods for deployment",
			"deployment", deploymentName,
		)
		return ctrl.Result{}, podListError
	}

	allUpdated := true
	for _, pod := range pods.Items {
		shouldUpdatePod := false
		if labelsUpdated := r.SetLabels(labelsToReflect, &pod); labelsUpdated {
			shouldUpdatePod = true
		}

		var labelsToUnset []string
		if value, annotationExists := pod.Annotations[ReflectorLabelsReflectedAnnotation]; annotationExists {
			currentKeys := strings.Split(value, ",")
			expectedKeys := common.MapKeysAsSlice(labelsToReflect)
			labelsToUnset = common.ExcessiveElements(expectedKeys, currentKeys)
		}
		if len(labelsToUnset) > 0 {
			if labelsUpdated := r.UnsetLabels(labelsToUnset, &pod); labelsUpdated {
				shouldUpdatePod = true
			}
		}

		if annotationsUpdated := r.SetAnnotations(reflectedAnnotations, &pod); annotationsUpdated {
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
			allUpdated = false
		}
	}

	if !allUpdated {
		r.logger.Info("Could not update all managed pods", "deployment", namespacedName)
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ReflectorController) UnsetReflectedLabels(ctx context.Context, deployment *appsv1.Deployment,
) (ctrl.Result, error) {
	deploymentName := deployment.Name

	pods, podListError := r.GetManagedPods(ctx, deployment)
	if podListError != nil {
		r.logger.Error(podListError,
			"Error listing pods for deployment",
			"deployment", deploymentName,
		)
		return ctrl.Result{}, podListError
	}

	allUpdated := true
	for _, pod := range pods.Items {
		annotationValue, hasReflectorAnnotation := pod.Annotations[ReflectorLabelsReflectedAnnotation]
		// if the annotation is not present, configuration is either already unset
		// or the annotation was deleted manually and we don't know what labels to remove
		if !hasReflectorAnnotation {
			continue
		}

		shouldUpdatePod := false

		labelsToUnset := strings.Split(annotationValue, ",")
		if labelsUpdated := r.UnsetLabels(labelsToUnset, &pod); labelsUpdated {
			shouldUpdatePod = true
		}

		annotationsToUnset := []string{ReflectorLabelsReflectedAnnotation}
		if annotationsUpdated := r.UnsetAnnotations(annotationsToUnset, &pod); annotationsUpdated {
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
			allUpdated = false
		}
	}

	if !allUpdated {
		r.logger.Info("Could not unset metadata, requeuing...", "deployment", deploymentName)
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// reflect a list of labels from deployment to the pod
// return whether any pod label was updated
func (r *ReflectorController) SetLabels(labels map[string]string, pod *v1.Pod) bool {
	podUpdated := false

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

// unset managed labels from a pod
// returns whether any label was unset
func (r *ReflectorController) UnsetLabels(labels []string, pod *v1.Pod) bool {
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

// given reflector annotations from the object and a map of labels, find labels that need to be reflected
func (r *ReflectorController) LabelsToReflect(
	reflectorAnnotations map[string]string, labels map[string]string,
) (map[string]string, error) {

	labelsToReflect := make(map[string]string)

	for annKey, annValue := range reflectorAnnotations {
		annotationKeyParts := strings.Split(annKey, "/")
		if len(annotationKeyParts) != 2 {
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
				if !regex.Match([]byte(key)) {
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
