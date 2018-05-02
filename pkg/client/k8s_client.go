package client

import (
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8sClint(host string) *k8s.Clientset {
	config, err := GetClientConfig(host)
	clientSet, err := k8s.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}

func GetClientConfig(host string) (*rest.Config, error) {
	if host != "" {
		return clientcmd.BuildConfigFromFlags(host, "")
	}
	return rest.InClusterConfig()
}
