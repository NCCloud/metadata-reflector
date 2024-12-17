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
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestController_reflectBackground(t *testing.T) {
	type args struct {
		labelSelector string
	}
	tests := []struct {
		name      string
		args      args
		mockSetup func(*mockKubernetesClient.MockKubernetesClient)
		wantErr   bool
	}{
		{
			name: "Successful background task",
			args: args{
				labelSelector: "",
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("ListDeployments", mock.Anything, mock.Anything).
					Return(&appsv1.DeploymentList{
						Items: []appsv1.Deployment{},
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "Invalid deployment selector",
			args: args{
				labelSelector: "invalid;",
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {},
			wantErr:   true,
		},
		{
			name: "Deployment list error",
			args: args{
				labelSelector: "",
			},
			mockSetup: func(mockClient *mockKubernetesClient.MockKubernetesClient) {
				mockClient.On("ListDeployments", mock.Anything, mock.Anything).
					Return(&appsv1.DeploymentList{}, errors.New("failed to list deployments"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockKubernetesClient.MockKubernetesClient)

			logger := zap.New()
			config := &common.Config{
				DeploymentSelector: tt.args.labelSelector,
			}

			controller := &Controller{
				kubeClient: mockClient,
				logger:     logger,
				config:     config,
			}
			tt.mockSetup(mockClient)

			err := controller.reflectBackground(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}
