package reflector

import (
	"context"
	"errors"
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

func TestController_reflectLabels(t *testing.T) {
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
			name: "Successfully reflect labels",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "label2",
						},
						Labels: map[string]string{
							"label1": "value1",
							"label2": "value2",
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
									Labels:      map[string]string{"label1": "value1"},
									Annotations: map[string]string{},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:        "pod2",
									Namespace:   "default",
									Labels:      map[string]string{"label1": "value1"},
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
			name: "Invalid annotation",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/invalid": "",
						},
						Labels: map[string]string{},
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {},
			want:      ctrl.Result{},
			wantErr:   true,
		},
		{
			name: "No labels to reflect",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "",
						},
						Labels: map[string]string{},
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
			name: "Cannot unset metadata",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "",
						},
						Labels: map[string]string{},
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
										"labels.metadata-reflector.spaceship.com/reflected-list": "label1",
									},
									Labels: map[string]string{
										"label1": "value1",
									},
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(errors.New("update failed"))
			},
			want:    ctrl.Result{},
			wantErr: true,
		},
		{
			name: "Error listing pods",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "key1",
						},
						Labels: map[string]string{
							"key1": "value1",
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
					Return(&v1.PodList{}, errors.New("failed to list pods"))
			},
			want:    ctrl.Result{},
			wantErr: true,
		},
		{
			name: "Could not update all managed pods",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "key1",
						},
						Labels: map[string]string{
							"key1": "value1",
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
									Name:      "pod1",
									Namespace: "default",
									Annotations: map[string]string{
										"labels.metadata-reflector.spaceship.com/reflected-list": "label1",
									},
									Labels: map[string]string{
										"label1": "value1",
									},
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(errors.New("update failed"))
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

			got, err := controller.reflectLabels(context.Background(), tt.args.deployment)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_unsetReflectedLabels(t *testing.T) {
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
			name: "Successfully unset labels",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test-deployment",
						Namespace:   "default",
						Annotations: map[string]string{},
						Labels:      map[string]string{},
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
									Labels:      map[string]string{"label1": "value1"},
									Annotations: map[string]string{},
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "pod2",
									Namespace: "default",
									Labels:    map[string]string{"label1": "value1"},
									Annotations: map[string]string{
										"labels.metadata-reflector.spaceship.com/reflected-list": "label1",
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
			name: "Cannot unset metadata",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "",
						},
						Labels: map[string]string{},
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
										"labels.metadata-reflector.spaceship.com/reflected-list": "label1",
									},
									Labels: map[string]string{
										"label1": "value1",
									},
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(errors.New("update failed"))
			},
			want:    ctrl.Result{},
			wantErr: true,
		},
		{
			name: "Error listing pods",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "key1",
						},
						Labels: map[string]string{
							"key1": "value1",
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
					Return(&v1.PodList{}, errors.New("failed to list pods"))
			},
			want:    ctrl.Result{},
			wantErr: true,
		},
		{
			name: "Could not update all managed pods",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
						Annotations: map[string]string{
							"labels.metadata-reflector.spaceship.com/list": "key1",
						},
						Labels: map[string]string{
							"key1": "value1",
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
									Name:      "pod1",
									Namespace: "default",
									Annotations: map[string]string{
										"labels.metadata-reflector.spaceship.com/reflected-list": "label1",
									},
									Labels: map[string]string{
										"label1": "value1",
									},
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(errors.New("update failed"))
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

			got, err := controller.reflectLabels(context.Background(), tt.args.deployment)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_setLabels(t *testing.T) {
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
			Labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	labelsToSet := map[string]string{
		"key1": "value1",     // existing label, should not be updated
		"key2": "new-value2", // label value changed, should be updated
		"key3": "value3",     // new label, should be added
	}

	updated := controller.setLabels(labelsToSet, pod)

	assert.True(t, updated, "The pod should be updated because labels were set.")
	assert.Equal(t, "new-value2", pod.Labels["key2"], "Label 'key2' should be updated.")
	assert.Equal(t, "value3", pod.Labels["key3"], "Label 'key3' should be added.")
}

func TestController_setLabelsToPodWithoutAnyExistingOnes(t *testing.T) {
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
		},
	}

	labelsToSet := map[string]string{
		"key1": "value1", // new label, should be added
		"key3": "value3", // new label, should be added
	}

	updated := controller.setLabels(labelsToSet, pod)

	assert.True(t, updated, "The pod should be updated because labels were set.")
	assert.Equal(t, "value1", pod.Labels["key1"], "Label 'key1' should be added.")
	assert.Equal(t, "value3", pod.Labels["key3"], "Label 'key3' should be added.")
}

func TestController_unsetLabels(t *testing.T) {
	mockKubernetesClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockKubernetesClient,
		logger:     logger,
		config:     config,
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	labelsToUnset := []string{
		"key2",
		"key3", // label not present in pod, so nothing should be deleted
	}

	anyLabelDeleted := controller.unsetLabels(labelsToUnset, pod)

	assert.True(t, anyLabelDeleted, "Some labels should be deleted.")
	assert.Equal(t, map[string]string{"key1": "value1"}, pod.Labels, "Pod labels should be different.")
}

func TestController_unsetLabelsFromPodWithoutAnyExistingOnes(t *testing.T) {
	mockKubernetesClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockKubernetesClient,
		logger:     logger,
		config:     config,
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	anyLabelDeleted := controller.unsetLabels(labelsToUnset, pod)

	assert.False(t, anyLabelDeleted, "No labels should be updated")
}

func TestController_unsetExcessiveLabels(t *testing.T) {
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
			Labels: map[string]string{
				"label1": "value1",
				"label2": "value2",
				"label3": "value3",
			},
			Annotations: map[string]string{
				ReflectorLabelsReflectedAnnotation: "label1,label2,label3",
			},
		},
	}

	labelsToReflect := map[string]string{
		"label1": "value1", // Expected label
		"label2": "value2", // Expected label
	}

	anyLabelDeleted := controller.unsetExcessiveLabels(labelsToReflect, pod)

	assert.True(t, anyLabelDeleted, "Some labels should be deleted.")
	assert.NotContains(t, pod.Labels, "label3", "Label 'label3' should be unset.")
	assert.Contains(t, pod.Labels, "label2", "Label 'label2' should remain.")
}

func TestController_labelsToReflect(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	type args struct {
		reflectorAnnotations map[string]string
		labels               map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Valid regex annotation",
			args: args{
				reflectorAnnotations: map[string]string{
					"reflector/regex": "^key[1-2]$",
				},
				labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "Valid list annotation",
			args: args{
				reflectorAnnotations: map[string]string{
					"reflector/list": "key1,key2",
				},
				labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name: "Invalid annotation",
			args: args{
				reflectorAnnotations: map[string]string{
					"invalid-annotation": "key1,key2",
				},
				labels: map[string]string{
					"key1": "value1",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Unsupported operation",
			args: args{
				reflectorAnnotations: map[string]string{
					"reflector/exclude": "key1",
				},
				labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := controller.labelsToReflect(tt.args.reflectorAnnotations, tt.args.labels)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
