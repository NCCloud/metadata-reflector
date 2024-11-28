package reflector

import "fmt"

var ReflectorAnnotationDomain = "metadata-reflector.spaceship.com"

// operations to get labels, e.g. `list`, `regex`, etc.
var (
	ReflectorLabelsList  = "list"
	ReflectorLabelsRegex = "regex"
)

var (
	ReflectorLabelsAnnotationDomain = fmt.Sprintf("labels.%s", ReflectorAnnotationDomain)

	ReflectorLabelsListAnnotation = fmt.Sprintf(
		"%s/%s", ReflectorLabelsAnnotationDomain, ReflectorLabelsList)

	ReflectorLabelsRegexAnnotation = fmt.Sprintf(
		"%s/%s", ReflectorLabelsAnnotationDomain, ReflectorLabelsRegex)

	// a list of annotations that were added to the object by the controller.
	ReflectorLabelsReflectedAnnotation = fmt.Sprintf("%s/%s", ReflectorLabelsAnnotationDomain, "reflected-list")
)
