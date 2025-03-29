package util

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
)

func GetPodAndEnvVarByContainerName(ctx context.Context, clientset *kubernetes.Clientset, namespace, containerName, envVarName string) (string, string, error) {
	podName, envVals, err := GetPodAndEnvVarsByContainerName(ctx, clientset, namespace, containerName, envVarName)
	if err != nil {
		return "", "", err
	}
	if len(envVals) > 0 {
		if val, ok := envVals[envVarName]; ok {
			return podName, val, nil
		}
	}
	return "", "", fmt.Errorf("pod or env var not found")
}

func GetPodAndEnvVarsByContainerName(ctx context.Context, clientset *kubernetes.Clientset, namespace, containerName string, envVarNames ...string) (string, map[string]string, error) {
	envVals := map[string]string{}
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", nil, err
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if container.Name == containerName {

				for _, env := range container.Env {
					for _, evn := range envVarNames {
						if env.Name == evn {
							envVals[env.Name] = env.Value
						}
					}
				}

				if len(envVals) > 0 {
					return pod.Name, envVals, nil
				} else {
					return "", nil, fmt.Errorf("pod or env vars not found")
				}

			}
		}
	}

	return "", nil, fmt.Errorf("pod or env vars not found")
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

func FileExists(fp string) bool {
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
