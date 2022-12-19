package kube

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

const SYSTEM_NS = "sreplatform-system"

var Flags = genericclioptions.NewConfigFlags(true)

func NewClient() (*kubernetes.Clientset, error) {
	cfg, err := Flags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	client, _ := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
