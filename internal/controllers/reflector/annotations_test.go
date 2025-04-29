package reflector

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/NCCloud/metadata-reflector/internal/common"
	mockKubernetesClient "github.com/NCCloud/metadata-reflector/mocks/github.com/NCCloud/metadata-reflector/internal_/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestController_reconcileAnnotations(t *testing.T) {
	type args struct {
		deployment *appsv1.Deployment
	}
	tests := []struct {
		name      string
		args      args
		mockSetup func(*mockKubernetesClient.MockKubernetesClient)
		want      ctrl.Result
		wantErr   bool
	}{
		{
			name: "Successful reconciliation: add new annotations",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							fmt.Sprintf("%s/list", ReflectorAnnotationsAnnotationDomain): "annotation1",
							"annotation1": "value1",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				// Mock managed pods
				mockClient.On("ListPods", mock.Anything, mock.Anything).
					Return(&v1.PodList{
						Items: []v1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:        "pod1",
									Namespace:   "default",
									Labels:      map[string]string{},
									Annotations: map[string]string{},
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(nil)
			},
			want:    ctrl.Result{},
			wantErr: false,
		},
		{
			name: "Successful reconciliation: remove annotations",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				// Mock managed pods
				mockClient.On("ListPods", mock.Anything, mock.Anything).
					Return(&v1.PodList{
						Items: []v1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "pod1",
									Namespace: "default",
									Annotations: map[string]string{
										fmt.Sprintf("%s/reflected-list", ReflectorAnnotationsAnnotationDomain): "annotation1",
										"annotation1": "value1",
									},
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(nil)
			},
			want:    ctrl.Result{},
			wantErr: false,
		},
		{
			name: "Failed reconciliation",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							fmt.Sprintf("%s/list", ReflectorAnnotationsAnnotationDomain): "annotation1",
							"annotation1": "value1",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				// Mock managed pods
				mockClient.On("ListPods", mock.Anything, mock.Anything).
					Return(&v1.PodList{
						Items: []v1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:        "pod1",
									Namespace:   "default",
									Labels:      map[string]string{},
									Annotations: map[string]string{},
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(errors.New("failed"))
			},
			want:    ctrl.Result{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockKubernetesClient.MockKubernetesClient)

			logger := zap.New()
			config := &common.Config{}

			controller := &Controller{
				kubeClient: mockClient,
				logger:     logger,
				config:     config,
			}
			tt.mockSetup(mockClient)

			got, err := controller.reconcileAnnotations(context.Background(), tt.args.deployment)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_getReflectorAnnForAnnotations(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	annotationsToReflect := "key1=value1,key2=value2"
	expectedAnnotations := map[string]string{
		ReflectorAnnotationsReflectedAnnotation: annotationsToReflect,
	}

	result := controller.getReflectorAnnForAnnotations(annotationsToReflect)

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

func TestController_unsetAnnotationsFromPodWithoutAnyExistingOnes(t *testing.T) {
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
		},
	}

	annotationsUnset := controller.unsetAnnotations(annotationsToUnset, pod)

	assert.False(t, annotationsUnset, "annotationsUnset should be false because there are no annotations to unset")
}

func TestController_unsetExcessiveAnnotations(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				ReflectorAnnotationsReflectedAnnotation: "annotation1,annotation2,annotation3",
				"annotation1":                           "value1",
				"annotation2":                           "value2",
				"annotation3":                           "value3",
			},
		},
	}

	expectedAnnotations := map[string]string{
		"annotation1": "value1", // Expected annotation
		"annotation2": "value2", // Expected annotation
	}

	anyAnnotationDeleted := controller.unsetExcessiveAnnotations(expectedAnnotations, pod)

	assert.True(t, anyAnnotationDeleted, "Some annotations should be deleted.")
	assert.NotContains(t, pod.Annotations, "annotation3", "Annotation 'annotation3' should be unset.")
	assert.Contains(t, pod.Annotations, "annotation2", "Label 'annotation2' should remain.")
}
