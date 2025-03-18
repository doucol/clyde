package util

import (
	"bytes"
	"context"
	"fmt"

	"github.com/doucol/clyde/internal/cmdContext"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func ReadFileFromPod(ctx context.Context, namespace, podName, containerName, filePath string) (string, error) {
	config := cmdContext.K8sConfigFromContext(ctx)
	clientset := cmdContext.ClientsetFromContext(ctx)
	// Prepare the request to execute command in the pod
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"/bin/sh", "-c", "cat " + filePath},
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, runtime.NewParameterCodec(scheme.Scheme))

		// Execute the request
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %v", err)
	}

	// Buffer to store the output
	var stdout, stderr bytes.Buffer

	// Stream options
	streamOptions := remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	// Execute the command and stream the output
	err = exec.StreamWithContext(ctx, streamOptions)
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("failed to execute command: %v, stderr: %s", err, stderr.String())
		}
		return "", fmt.Errorf("failed to execute command: %v", err)
	}

	// Check if there is any error in stderr
	if stderr.Len() > 0 {
		return "", fmt.Errorf("command executed with error: %s", stderr.String())
	}

	return stdout.String(), nil
}
