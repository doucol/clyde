package util

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/util/homedir"
)

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	if err != nil {
		t.Errorf("GetFreePort() error = %v", err)
	}
	if port <= 0 {
		t.Errorf("GetFreePort() returned invalid port: %d", port)
	}

	// Test that the port is actually free
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Errorf("Failed to resolve TCP addr: %v", err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Errorf("Port %d is not free: %v", port, err)
	}
	listener.Close()
}

func TestGetDataPath(t *testing.T) {
	// Test with default XDG_DATA_HOME
	os.Unsetenv("XDG_DATA_HOME")
	path := GetDataPath()
	expectedPath := filepath.Join(homedir.HomeDir(), ".local", "share", "clyde")
	if path != expectedPath {
		t.Errorf("GetDataPath() = %v; want %v", path, expectedPath)
	}

	// Test with custom XDG_DATA_HOME

	customPath := filepath.Join(os.TempDir(), "custom-data-path")
	os.Setenv("XDG_DATA_HOME", customPath)
	defer os.Unsetenv("XDG_DATA_HOME")

	path = GetDataPath()
	expectedPath = filepath.Join(customPath, "clyde")
	if path != expectedPath {
		t.Errorf("GetDataPath() with custom XDG_DATA_HOME = %v; want %v", path, expectedPath)
	}

	// Verify directory was created
	if !FileExists(path) {
		t.Errorf("GetDataPath() did not create directory at %v", path)
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Test existing file
	if !FileExists(tmpFile.Name()) {
		t.Errorf("FileExists() = false for existing file %v", tmpFile.Name())
	}

	// Test non-existing file
	nonExistentFile := filepath.Join(os.TempDir(), "non-existent-file")
	if FileExists(nonExistentFile) {
		t.Errorf("FileExists() = true for non-existent file %v", nonExistentFile)
	}
}

func TestGetPodAndEnvVarsByContainerName(t *testing.T) {
	// Create a fake Kubernetes client
	clientset := fake.NewSimpleClientset()

	// Create test pod with environment variables
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "test-container",
					Env: []corev1.EnvVar{
						{
							Name:  "TEST_VAR1",
							Value: "value1",
						},
						{
							Name:  "TEST_VAR2",
							Value: "value2",
						},
					},
				},
			},
		},
	}

	// Add the pod to the fake client
	_, err := clientset.CoreV1().Pods("test-namespace").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test pod: %v", err)
	}

	tests := []struct {
		name          string
		namespace     string
		containerName string
		envVarNames   []string
		wantPod       string
		wantEnvVars   map[string]string
		wantErr       bool
	}{
		{
			name:          "existing pod and env vars",
			namespace:     "test-namespace",
			containerName: "test-container",
			envVarNames:   []string{"TEST_VAR1", "TEST_VAR2"},
			wantPod:       "test-pod",
			wantEnvVars: map[string]string{
				"TEST_VAR1": "value1",
				"TEST_VAR2": "value2",
			},
			wantErr: false,
		},
		{
			name:          "non-existent namespace",
			namespace:     "non-existent",
			containerName: "test-container",
			envVarNames:   []string{"TEST_VAR1"},
			wantPod:       "",
			wantEnvVars:   nil,
			wantErr:       true,
		},
		{
			name:          "non-existent container",
			namespace:     "test-namespace",
			containerName: "non-existent",
			envVarNames:   []string{"TEST_VAR1"},
			wantPod:       "",
			wantEnvVars:   nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPod, gotEnvVars, err := GetPodAndEnvVarsByContainerName(context.Background(), clientset, tt.namespace, tt.containerName, tt.envVarNames...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodAndEnvVarsByContainerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPod != tt.wantPod {
				t.Errorf("GetPodAndEnvVarsByContainerName() pod = %v, want %v", gotPod, tt.wantPod)
			}
			if !tt.wantErr {
				for k, v := range tt.wantEnvVars {
					if gotEnvVars[k] != v {
						t.Errorf("GetPodAndEnvVarsByContainerName() envVars[%v] = %v, want %v", k, gotEnvVars[k], v)
					}
				}
			}
		})
	}
}

func TestGetPodAndEnvVarByContainerName(t *testing.T) {
	// Create a fake Kubernetes client
	clientset := fake.NewSimpleClientset()

	// Create test pod with environment variables
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "test-container",
					Env: []corev1.EnvVar{
						{
							Name:  "TEST_VAR",
							Value: "test-value",
						},
					},
				},
			},
		},
	}

	// Add the pod to the fake client
	_, err := clientset.CoreV1().Pods("test-namespace").Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test pod: %v", err)
	}

	tests := []struct {
		name          string
		namespace     string
		containerName string
		envVarName    string
		wantPod       string
		wantValue     string
		wantErr       bool
	}{
		{
			name:          "existing pod and env var",
			namespace:     "test-namespace",
			containerName: "test-container",
			envVarName:    "TEST_VAR",
			wantPod:       "test-pod",
			wantValue:     "test-value",
			wantErr:       false,
		},
		{
			name:          "non-existent namespace",
			namespace:     "non-existent",
			containerName: "test-container",
			envVarName:    "TEST_VAR",
			wantPod:       "",
			wantValue:     "",
			wantErr:       true,
		},
		{
			name:          "non-existent container",
			namespace:     "test-namespace",
			containerName: "non-existent",
			envVarName:    "TEST_VAR",
			wantPod:       "",
			wantValue:     "",
			wantErr:       true,
		},
		{
			name:          "non-existent env var",
			namespace:     "test-namespace",
			containerName: "test-container",
			envVarName:    "NON_EXISTENT_VAR",
			wantPod:       "",
			wantValue:     "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPod, gotValue, err := GetPodAndEnvVarByContainerName(context.Background(), clientset, tt.namespace, tt.containerName, tt.envVarName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodAndEnvVarByContainerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPod != tt.wantPod {
				t.Errorf("GetPodAndEnvVarByContainerName() pod = %v, want %v", gotPod, tt.wantPod)
			}
			if gotValue != tt.wantValue {
				t.Errorf("GetPodAndEnvVarByContainerName() value = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}
