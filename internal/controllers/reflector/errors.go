package reflector

import "errors"

var (
	ErrUnparsableAnnotation = errors.New("annotation cannot be parsed")
	ErrUnparsableOperation  = errors.New("operation cannot be parsed")
	ErrEmptyPodSelector     = errors.New("empty pod selector")
	ErrPodNotFound          = errors.New("failed to find pods")
	ErrPodsUpdateFailed     = errors.New("failed to update pods")
)
