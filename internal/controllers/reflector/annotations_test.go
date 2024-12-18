package reflector

import (
	"testing"

	"github.com/NCCloud/metadata-reflector/internal/common"
	mockKubernetesClient "github.com/NCCloud/metadata-reflector/mocks/github.com/NCCloud/metadata-reflector/internal_/clients"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestController_getReflectedAnnotations(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	labelsToReflect := "key1=value1,key2=value2"
	expectedAnnotations := map[string]string{
		ReflectorLabelsReflectedAnnotation: labelsToReflect,
	}

	result := controller.getReflectedAnnotations(labelsToReflect)

	assert.NotNil(t, result, "resulting annotations should not be nil")
	assert.Equal(t, expectedAnnotations, result, "annotations should match the expected map")
}

func TestController_setAnnotations(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	annotationsToSet := map[string]string{
		"annotation1": "value1",
		"annotation2": "value2",
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-pod",
			Annotations: map[string]string{"annotation1": "oldValue", "annotation3": "value"},
		},
	}

	expectedAnnotations := map[string]string{
		"annotation1": "value1",
		"annotation2": "value2",
		"annotation3": "value",
	}

	podUpdated := controller.setAnnotations(annotationsToSet, pod)

	assert.True(t, podUpdated, "pod annotations should be updated")
	assert.Equal(t, expectedAnnotations, pod.Annotations, "pod annotations should match the expected annotations")
}

func TestController_setAnnotationsToPodWithoutAnyExistingOnes(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	annotationsToSet := map[string]string{
		"annotation1": "value1",
		"annotation2": "value2",
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
	}

	expectedAnnotations := map[string]string{
		"annotation1": "value1",
		"annotation2": "value2",
	}

	podUpdated := controller.setAnnotations(annotationsToSet, pod)

	assert.True(t, podUpdated, "pod with no existing annotations should be updated")
	assert.Equal(t, expectedAnnotations, pod.Annotations, "pod annotations should match the expected annotations")
}

func TestController_unsetAnnotations(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	annotationsToUnset := []string{"annotation1", "annotation3"}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"annotation1": "value1",
				"annotation2": "value2",
			},
		},
	}

	expectedAnnotations := map[string]string{
		"annotation2": "value2", // Only this annotation should remain
	}

	annotationsUnset := controller.unsetAnnotations(annotationsToUnset, pod)

	assert.True(t, annotationsUnset, "annotationsUnset should be true because annotation1 was removed")
	assert.Equal(t, expectedAnnotations, pod.Annotations, "remaining annotations should match the expected annotations")
}
