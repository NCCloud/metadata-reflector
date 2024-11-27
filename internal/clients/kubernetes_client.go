package clients

import (
	"context"

	"github.com/NCCloud/metadata-reflector/internal/common"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type KubernetesClient interface {
	ListDeployments(ctx context.Context, labelSelector labels.Selector) (*appsv1.DeploymentList, error)
	ListPods(ctx context.Context, labelSelector labels.Selector) (*v1.PodList, error)
	GetDeployment(ctx context.Context, namespacedName types.NamespacedName) (*appsv1.Deployment, error)
	UpdatePod(ctx context.Context, pod v1.Pod) error
}

type kubernetesClient struct {
	// used for read operations
	cacheClient cache.Cache
	// used for write operations
	client client.Client
	config *common.Config
}

func NewKubernetesClient(
	ctx context.Context, mgr manager.Manager, config *common.Config, logger logr.Logger,
) KubernetesClient {
	client := mgr.GetClient()
	cacheClient := mgr.GetCache()

	return &kubernetesClient{
		cacheClient: cacheClient,
		client:      client,
		config:      config,
	}
}

func (c *kubernetesClient) ListDeployments(ctx context.Context, labelSelector labels.Selector,
) (*appsv1.DeploymentList, error) {
	deploymentList := &appsv1.DeploymentList{}

	listOptions := &client.ListOptions{
		LabelSelector: labelSelector,
	}

	if listErr := c.cacheClient.List(ctx, deploymentList, listOptions); listErr != nil {
		return nil, listErr
	}

	return deploymentList, nil
}

func (c *kubernetesClient) ListPods(ctx context.Context, labelSelector labels.Selector,
) (*v1.PodList, error) {
	podList := &v1.PodList{}
	listOptions := &client.ListOptions{
		LabelSelector: labelSelector,
	}
	if listErr := c.cacheClient.List(ctx, podList, listOptions); listErr != nil {
		return nil, listErr
	}

	return podList, nil
}

func (c *kubernetesClient) GetDeployment(ctx context.Context, namespacedName types.NamespacedName,
) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	if getErr := c.cacheClient.Get(ctx, namespacedName, deployment); getErr != nil {
		return nil, getErr
	}

	return deployment, nil
}

func (c *kubernetesClient) UpdatePod(ctx context.Context, pod v1.Pod) error {
	return c.client.Update(ctx, &pod)
}
