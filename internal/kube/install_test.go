package kube

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynfake "k8s.io/client-go/dynamic/fake"
)

func TestCanInstallWhisker(t *testing.T) {
	tests := []struct {
		name string
		info ClusterNetworkingInfo
		want bool
	}{
		{
			name: "operator installed",
			info: ClusterNetworkingInfo{OperatorInstalled: true},
			want: true,
		},
		{
			name: "operator not installed",
			info: ClusterNetworkingInfo{OperatorInstalled: false},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanInstallWhisker(tt.info); got != tt.want {
				t.Errorf("CanInstallWhisker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstallGoldmaneWhisker(t *testing.T) {
	ctx := context.Background()
	dyn := dynfake.NewSimpleDynamicClient(runtime.NewScheme())

	if err := InstallGoldmaneWhisker(ctx, dyn); err != nil {
		t.Fatalf("InstallGoldmaneWhisker() unexpected error: %v", err)
	}

	// Both resources should now exist.
	if _, err := dyn.Resource(goldmaneGVR).Get(ctx, "default", metav1.GetOptions{}); err != nil {
		t.Errorf("expected Goldmane/default to exist: %v", err)
	}
	if _, err := dyn.Resource(whiskerGVR).Get(ctx, "default", metav1.GetOptions{}); err != nil {
		t.Errorf("expected Whisker/default to exist: %v", err)
	}

	// Calling again must be idempotent (AlreadyExists is tolerated).
	if err := InstallGoldmaneWhisker(ctx, dyn); err != nil {
		t.Errorf("InstallGoldmaneWhisker() second call should be idempotent, got: %v", err)
	}
}
