package reflector

import (
	"regexp"
	"slices"
	"strings"

	"github.com/NCCloud/metadata-reflector/internal/common"
)

func (r *Controller) validateAnnotation(annotation string) error {
	expectedAnnotationParts := 2
	annotationKeyParts := strings.Split(annotation, "/")
	if len(annotationKeyParts) != expectedAnnotationParts {
		r.logger.Error(ErrUnparsableAnnotation,
			"Annotation should consist of exactly 2 parts",
			"annotation", annotation, "parts", len(annotationKeyParts),
		)

		return ErrUnparsableAnnotation
	}

	annDomains := supportedAnnotationDomains()

	if !slices.Contains(annDomains, annotationKeyParts[0]) {
		r.logger.Error(ErrUnparsableAnnotation,
			"Annotation does not contain a valid annotation domain",
			"annotation", annotation, "supportedOperations", annDomains,
		)
		return ErrUnparsableAnnotation
	}

	operations := supportedOperations()

	if !slices.Contains(operations, annotationKeyParts[1]) {
		r.logger.Error(ErrUnparsableOperation,
			"Annotation does not contain a valid operation",
			"annotation", annotation, "supportedOperations", operations,
		)
		return ErrUnparsableOperation
	}

	return nil
}

// given a slice of reflector operations and a map of key-value pairs,
// find key-value paris that need to be reflected
func (r *Controller) keysToReflect(
	reflectorAnnotations map[string]string, data map[string]string,
) (map[string]string, error) {
	keysToReflect := make(map[string]string)

	for annKey, annValue := range reflectorAnnotations {
		annValidationErr := r.validateAnnotation(annKey)
		if annValidationErr != nil {
			return nil, annValidationErr
		}

		annotationKeyParts := strings.Split(annKey, "/")

		switch operation := annotationKeyParts[1]; operation {
		case ReflectorOperationList:
			annotationLabels := strings.Split(annValue, ",")

			for key, value := range data {
				if !slices.Contains(annotationLabels, key) {
					continue
				}

				keysToReflect[key] = value
			}

		case ReflectorOperationRegex:
			pattern := common.ExactMatchRegex(annValue)
			regex := regexp.MustCompile(pattern)

			for key, value := range data {
				if !regex.MatchString(key) {
					continue
				}

				keysToReflect[key] = value
			}

		default:
			r.logger.Error(ErrUnparsableOperation,
				"Annotation doesn't have a valid operation to parse", "annotation", annKey)

			return nil, ErrUnparsableOperation
		}
	}

	return keysToReflect, nil
}
