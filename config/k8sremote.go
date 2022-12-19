package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgmorales/go-cli-selfupdate/kube"
	"github.com/mitchellh/mapstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const defSSCfgNs = "myk8sapi-system"
const defSSCfgMap = "cli-config"

// KubeServerSideConfigLoader loads ServerSideConfig from a Config Map in a K8S Cluster
type KubeServerSideConfigLoader struct {
	client *kubernetes.Clientset
	ns     string
	cmName string
}

// NewKubeSSCfgLoader returns a KubeServerSideConfigLoader.
//
// If kc argument is nil, a new kubernetes.Clientset will be generated.
//
// If any of ns and cmName are empty (""), defaults are used for them.
func NewKubeSSCfgLoader(kc *kubernetes.Clientset, ns string, cmName string) (KubeServerSideConfigLoader, error) {
	if strings.TrimSpace(ns) == "" {
		ns = defSSCfgNs
	}

	if strings.TrimSpace(cmName) == "" {
		cmName = defSSCfgMap
	}

	if kc == nil {
		newKc, err := kube.NewClient()
		if err != nil {
			return KubeServerSideConfigLoader{}, err
		}

		return KubeServerSideConfigLoader{
			client: newKc,
			ns:     ns,
			cmName: cmName,
		}, nil
	}

	return KubeServerSideConfigLoader{
		client: kc,
		ns:     ns,
		cmName: cmName,
	}, nil
}

// Load loads server side configuration from a Kubernetes ConfigMap
func (k KubeServerSideConfigLoader) Load() (ServerSideConfig, error) {
	cfg := ServerSideConfig{}

	cm, err := k.client.CoreV1().ConfigMaps(k.ns).Get(context.Background(), k.cmName,
		metav1.GetOptions{})
	if err != nil {
		return cfg, fmt.Errorf("error reading configmap %s/%s: %w", k.ns, k.cmName, err)
	}

	err = mapstructure.Decode(cm.Data, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("error parsing configmap %s/%s: %w", k.ns, k.cmName, err)
	}

	return cfg, nil
}
