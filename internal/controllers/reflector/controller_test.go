package reflector

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NCCloud/metadata-reflector/internal/common"
	mockKubernetesClient "github.com/NCCloud/metadata-reflector/mocks/github.com/NCCloud/metadata-reflector/internal_/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestNewController(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := NewController(mockClient, logger, config)

	assert.NotNil(t, controller)
}

func TestController_Reconcile(t *testing.T) {
	type args struct {
		req ctrl.Request
	}
	tests := []struct {
		name      string
		args      args
		mockSetup func(*mockKubernetesClient.MockKubernetesClient)
		want      ctrl.Result
		wantErr   bool
	}{
		{
			name: "Successful reconciliation with label reflection",
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "default",
						Name:      "test-deployment",
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("GetDeployment", mock.Anything, mock.Anything).
					Return(&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-deployment",
							Namespace: "default",
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "label1",
							},
							Labels: map[string]string{
								"label1": "value1",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "test"},
							},
						},
					}, nil)

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
			name: "Successful reconciliation with label removal",
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "default",
						Name:      "test-deployment",
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("GetDeployment", mock.Anything, mock.Anything).
					Return(&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-deployment",
							Namespace: "default",
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "",
							},
							Labels: map[string]string{
								"label1": "value1",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "test"},
							},
						},
					}, nil)

				// Mock managed pods
				mockClient.On("ListPods", mock.Anything, mock.Anything).
					Return(&v1.PodList{
						Items: []v1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "pod1",
									Namespace: "default",
									Labels: map[string]string{
										"label1": "value1",
									},
									Annotations: map[string]string{
										fmt.Sprintf("%s/reflected-list", ReflectorLabelsAnnotationDomain): "label1",
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
			name: "Failed to get deployment",
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "default",
						Name:      "test-deployment",
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("GetDeployment", mock.Anything, mock.Anything).
					Return(&appsv1.Deployment{}, errors.New("failed to get deployment"))
			},
			want:    ctrl.Result{},
			wantErr: true,
		},
		{
			name: "Successful reconciliation with annotation reflection",
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "default",
						Name:      "test-deployment",
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("GetDeployment", mock.Anything, mock.Anything).
					Return(&appsv1.Deployment{
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
					}, nil)

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
			name: "Successful reconciliation with annotation removal",
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "default",
						Name:      "test-deployment",
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("GetDeployment", mock.Anything, mock.Anything).
					Return(&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-deployment",
							Namespace: "default",
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorAnnotationsAnnotationDomain): "",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "test"},
							},
						},
					}, nil)

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
			name: "Failed reconciliation of annotation reflection",
			args: args{
				req: ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: "default",
						Name:      "test-deployment",
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("GetDeployment", mock.Anything, mock.Anything).
					Return(&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-deployment",
							Namespace: "default",
							Annotations: map[string]string{
								fmt.Sprintf("%s/invalid", ReflectorAnnotationsAnnotationDomain): "annotation1",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "test"},
							},
						},
					}, nil)

				// Mock managed pods
				mockClient.On("ListPods", mock.Anything, mock.Anything).
					Return(&v1.PodList{
						Items: []v1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "pod1",
									Namespace: "default",
								},
							},
						},
					}, nil)

				// Mock pod updates
				mockClient.On("UpdatePod", mock.Anything, mock.Anything).
					Return(nil)
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

			got, err := controller.Reconcile(context.Background(), tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_getManagedPods(t *testing.T) {
	type args struct {
		ctx        context.Context
		deployment *appsv1.Deployment
	}
	tests := []struct {
		name      string
		args      args
		mockSetup func(*mockKubernetesClient.MockKubernetesClient)
		want      *v1.PodList
		wantErr   bool
	}{
		{
			name: "Successfully get managed pods",
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
						},
					}, nil)
			},
			want: &v1.PodList{
				Items: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "pod1",
							Namespace:   "default",
							Labels:      map[string]string{"label1": "value1"},
							Annotations: map[string]string{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty pod selector",
			args: args{
				deployment: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{},
					},
				},
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {},
			want:      nil,
			wantErr:   true,
		},
		{
			name: "Found no pods",
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
				mockClient.On("ListPods", mock.Anything, mock.Anything).
					Return(&v1.PodList{Items: []v1.Pod{}}, nil)
			},
			want:    nil,
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

			got, err := controller.getManagedPods(context.Background(), tt.args.deployment)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_FilterCreateEvents(t *testing.T) {
	type args struct {
		e event.CreateEvent
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Event is not a Deployment",
			args: args{
				e: event.CreateEvent{
					Object: &v1.Pod{},
				},
			},
			want: false,
		},
		{
			name: "Event contains reflector annotation",
			args: args{
				e: event.CreateEvent{
					Object: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "key",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Event does not contain reflector annotation",
			args: args{
				e: event.CreateEvent{
					Object: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{},
						},
					},
				},
			},
			want: false,
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

			got := controller.FilterCreateEvents(tt.args.e)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_FilterUpdateEvents(t *testing.T) {
	type args struct {
		e event.UpdateEvent
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "New object is not a Deployment",
			args: args{
				e: event.UpdateEvent{
					ObjectNew: &v1.Pod{},
					ObjectOld: &appsv1.Deployment{},
				},
			},
			want: false,
		},
		{
			name: "Event is not a Deployment (old object)",
			args: args{
				e: event.UpdateEvent{
					ObjectNew: &appsv1.Deployment{},
					ObjectOld: &v1.Pod{}, // Not a Deployment
				},
			},
			want: false,
		},
		{
			name: "Neither old nor new deployment has reflector annotation",
			args: args{
				e: event.UpdateEvent{
					ObjectNew: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{},
						},
					},
					ObjectOld: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Deployment annotations changed",
			args: args{
				e: event.UpdateEvent{
					ObjectNew: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "new-value",
							},
						},
					},
					ObjectOld: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "old-value",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Deployment scaled up",
			args: args{
				e: event.UpdateEvent{
					ObjectNew: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "value",
							},
						},
						Status: appsv1.DeploymentStatus{
							ReadyReplicas: 1,
						},
					},
					ObjectOld: &appsv1.Deployment{
						Status: appsv1.DeploymentStatus{
							ReadyReplicas: 0,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Deployment labels changed",
			args: args{
				e: event.UpdateEvent{
					ObjectNew: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"new-label": "value",
							},
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "value",
							},
						},
					},
					ObjectOld: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"old-label": "value",
							},
							Annotations: map[string]string{
								fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "value",
							},
						},
					},
				},
			},
			want: true,
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

			got := controller.FilterUpdateEvents(tt.args.e)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_shouldRequeueNow(t *testing.T) {
	type args struct {
		result ctrl.Result
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Requeue now",
			args: args{
				result: ctrl.Result{RequeueAfter: 1 * time.Second},
			},
			want: true,
		},
		{
			name: "Delay requeue",
			args: args{
				result: ctrl.Result{},
			},
			want: false,
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

			got := controller.shouldRequeueNow(tt.args.result)

			assert.Equal(t, tt.want, got)
		})
	}
}
