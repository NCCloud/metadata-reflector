package clients

import (
	"context"
	"fmt"
	"testing"

	"github.com/NCCloud/metadata-reflector/internal/common"
	mockCache "github.com/NCCloud/metadata-reflector/mocks/sigs.k8s.io/controller-runtime/pkg/cache"
	mockClient "github.com/NCCloud/metadata-reflector/mocks/sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	_ "sigs.k8s.io/controller-runtime/pkg/cache"
	realClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNewKubernetesClient(t *testing.T) {
	config := common.NewConfig()
	mgr, _ := ctrl.NewManager(&rest.Config{}, ctrl.Options{})

	client := NewKubernetesClient(mgr, config)

	assert.NotNil(t, client)
}

func TestKubernetesClient_ListDeployments(t *testing.T) {
	ctx := context.Background()
	config := common.NewConfig()
	mockCache := new(mockCache.MockCache)
	mockClient := new(mockClient.MockClient)

	labelSelector, _ := labels.Parse("matching=true")

	matchingDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "matching-deployment",
			Labels: map[string]string{"matching": "true"},
		},
	}

	notMatchingDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "not-matching-deployment",
			Labels: map[string]string{"hello": "world"},
		},
	}

	expectedDeploymentList := &appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			matchingDeployment,
		},
	}

	deploymentList := []appsv1.Deployment{
		matchingDeployment,
		notMatchingDeployment,
	}

	// simulate filtering logic in the mock
	mockCache.On("List", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, list realClient.ObjectList, opts ...realClient.ListOption) error {
			if depList, ok := list.(*appsv1.DeploymentList); ok {
				for _, deployment := range deploymentList {
					if labelSelector.Matches(labels.Set(deployment.Labels)) {
						depList.Items = append(depList.Items, deployment)
					}
				}
			}
			return nil
		})

	client := &kubernetesClient{
		cacheClient: mockCache,
		client:      mockClient,
		config:      config,
	}

	result, listErr := client.ListDeployments(ctx, labelSelector)

	fmt.Println(len(result.Items), labelSelector.String())

	assert.Nil(t, listErr)
	assert.NotNil(t, result)
	assert.Equal(t, expectedDeploymentList, result)

	mockCache.AssertExpectations(t)
}

func TestKubernetesClient_ListPods(t *testing.T) {
	ctx := context.Background()
	config := common.NewConfig()
	mockCache := new(mockCache.MockCache)
	mockClient := new(mockClient.MockClient)

	labelSelector, _ := labels.Parse("matching=true")

	matchingPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "matching-pod",
			Labels: map[string]string{"matching": "true"},
		},
	}

	notMatchingPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "not-matching-pod",
			Labels: map[string]string{"hello": "world"},
		},
	}

	expectedPodList := &v1.PodList{
		Items: []v1.Pod{
			matchingPod,
		},
	}

	allPodList := []v1.Pod{
		matchingPod,
		notMatchingPod,
	}

	mockCache.On("List", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, list realClient.ObjectList, opts ...realClient.ListOption) error {
			if podList, ok := list.(*v1.PodList); ok {
				for _, pod := range allPodList {
					if labelSelector.Matches(labels.Set(pod.Labels)) {
						podList.Items = append(podList.Items, pod)
					}
				}
			}
			return nil
		})

	client := &kubernetesClient{
		cacheClient: mockCache,
		client:      mockClient,
		config:      config,
	}

	result, listErr := client.ListPods(ctx, labelSelector)

	fmt.Println(len(result.Items), labelSelector.String())

	assert.Nil(t, listErr)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPodList, result)

	mockCache.AssertExpectations(t)
}

func TestKubernetesClient_GetDeployment(t *testing.T) {
	ctx := context.Background()
	config := common.NewConfig()
	mockCache := new(mockCache.MockCache)
	mockClient := new(mockClient.MockClient)

	expectedDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
	}

	namespacedName := types.NamespacedName{
		Name:      "test-deployment",
		Namespace: "default",
	}

	objectKey := realClient.ObjectKey{
		Name:      namespacedName.Name,
		Namespace: namespacedName.Namespace,
	}

	mockCache.On("Get", mock.Anything, objectKey, mock.AnythingOfType("*v1.Deployment")).
		Run(func(args mock.Arguments) {
			// Populate the deployment object as if it were fetched from the cache
			if dep, ok := args.Get(2).(*appsv1.Deployment); ok {
				*dep = *expectedDeployment
			}
		}).
		Return(nil)

	client := &kubernetesClient{
		cacheClient: mockCache,
		client:      mockClient,
		config:      config,
	}

	result, getErr := client.GetDeployment(ctx, namespacedName)

	assert.Nil(t, getErr)
	assert.NotNil(t, result)
	assert.Equal(t, expectedDeployment, result)

	mockCache.AssertExpectations(t)
}

func TestKubernetesClient_UpdatePod(t *testing.T) {
	ctx := context.Background()
	mockClient := new(mockClient.MockClient)

	podToUpdate := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	}

	mockClient.On("Update", mock.Anything, mock.AnythingOfType("*v1.Pod")).
		Run(func(args mock.Arguments) {
			if pod, ok := args.Get(1).(*v1.Pod); ok {
				pod.Labels = map[string]string{"new-label": "true"} // Example modification
			}
		}).
		Return(nil)

	client := &kubernetesClient{
		client: mockClient,
	}

	updateErr := client.UpdatePod(ctx, podToUpdate)

	assert.Nil(t, updateErr)

	mockClient.AssertExpectations(t)
}
