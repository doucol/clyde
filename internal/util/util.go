package util

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/doucol/clyde/internal/cmdContext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
)

func GetPodAndEnvVarWithContainerName(ctx context.Context, namespace string, containerName string, envVarName string) (string, string, error) {
	clientset := cmdContext.ClientsetFromContext(ctx)
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", "", err
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if container.Name == containerName {
				for _, env := range container.Env {
					if env.Name == envVarName {
						return pod.Name, env.Value, nil
					}
				}
			}
		}
	}

	return "", "", fmt.Errorf("pod or env var name not found")
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func GetDataPath() string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(homedir.HomeDir(), ".local", "share")
	}
	dataDir = filepath.Join(dataDir, "clyde")
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		panic(err)
	}
	return dataDir
}
