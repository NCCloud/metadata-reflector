package reflector

import "errors"

var ErrUnparsableAnnotation = errors.New("Annotation cannot be parsed")
var ErrEmptyPodSelector = errors.New("Empty pod selector")
