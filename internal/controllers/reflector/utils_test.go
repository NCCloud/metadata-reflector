package reflector

import (
	"fmt"
	"testing"

	"github.com/NCCloud/metadata-reflector/internal/common"
	mockKubernetesClient "github.com/NCCloud/metadata-reflector/mocks/github.com/NCCloud/metadata-reflector/internal_/clients"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestController_validateAnnotation(t *testing.T) {
	mockClient := new(mockKubernetesClient.MockKubernetesClient)

	logger := zap.New()
	config := &common.Config{}

	controller := &Controller{
		kubeClient: mockClient,
		logger:     logger,
		config:     config,
	}

	type args struct {
		annotation string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Valid annotation key",
			args: args{
				annotation: fmt.Sprintf("%s/regex", ReflectorLabelsAnnotationDomain),
			},
			wantErr: false,
		},
		{
			name: "Invalid annotation key",
			args: args{
				annotation: "invalid-annotation-regex",
			},
			wantErr: true,
		},
		{
			name: "Invalid operation",
			args: args{
				annotation: fmt.Sprintf("%s/invalid", ReflectorLabelsAnnotationDomain),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := controller.validateAnnotation(tt.args.annotation)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, nil, err)
			}
		})
	}
}

func TestController_keysToReflect(t *testing.T) {
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
					fmt.Sprintf("%s/regex", ReflectorLabelsAnnotationDomain): "^key[1-2]$",
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
					fmt.Sprintf("%s/list", ReflectorLabelsAnnotationDomain): "key1,key2",
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
					fmt.Sprintf("%s/exclude", ReflectorLabelsAnnotationDomain): "key1",
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
			got, err := controller.keysToReflect(tt.args.reflectorAnnotations, tt.args.labels)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
