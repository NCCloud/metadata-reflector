package reflector

import "errors"

var (
	ErrUnparsableAnnotation = errors.New("annotation cannot be parsed")
	ErrEmptyPodSelector     = errors.New("empty pod selector")
	ErrPodNotFound          = errors.New("failed to find pods")
)
