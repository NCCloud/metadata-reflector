package reflector

import (
	v1 "k8s.io/api/core/v1"
)

// generate a map of annotations that help identify what was set by the reflector on the managed object
func (r *ReflectorController) GetReflectedAnnotations(labelsToReflect string) map[string]string {
	annotationsToReflect := make(map[string]string)
	annotationsToReflect[ReflectorLabelsReflectedAnnotation] = labelsToReflect

	return annotationsToReflect
}

// reflect annotation to managed pods
// returns whether any pod annotation was updated
func (r *ReflectorController) SetAnnotations(annotations map[string]string, pod *v1.Pod) bool {
	podUpdated := false

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

// unset managed annotations from a pod
// returns whether any annotation was unset
func (r *ReflectorController) UnsetAnnotations(annotations []string, pod *v1.Pod) bool {
	anyAnnotationUnset := false
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
