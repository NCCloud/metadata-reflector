package reflector

import "fmt"

var ReflectorAnnotationDomain = "metadata-reflector.spaceship.com"

// operations to get labels & annotations, e.g. `list`, `regex`, etc.
var (
	ReflectorOperationList  = "list"
	ReflectorOperationRegex = "regex"
)

var (
	ReflectorLabelsAnnotationDomain      = fmt.Sprintf("labels.%s", ReflectorAnnotationDomain)
	ReflectorAnnotationsAnnotationDomain = fmt.Sprintf("annotations.%s", ReflectorAnnotationDomain)

	// a list of annotations that were added to the object by the controller.
	ReflectorLabelsReflectedAnnotation = fmt.Sprintf("%s/%s", ReflectorLabelsAnnotationDomain, "reflected-list")

	ReflectorAnnotationsReflectedAnnotation = fmt.Sprintf(
		"%s/%s", ReflectorAnnotationsAnnotationDomain, "reflected-list")
)

func supportedAnnotationDomains() []string {
	return []string{ReflectorLabelsAnnotationDomain, ReflectorAnnotationsAnnotationDomain}
}

func supportedOperations() []string {
	return []string{ReflectorOperationList, ReflectorOperationRegex}
}
