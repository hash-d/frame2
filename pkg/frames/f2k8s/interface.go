package k8svalidate

import (
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Namespacer interface {
	GetNamespace() string
	Kube() configv1.KubeClientConfig
}

// This is to be used as an embedded field; implement the whole of
// Namespacer here
//
// Consider this: the PodInterface below (and others) might be interesting,
// but they'd pollute the frame's namespace if embedded
type Namespace struct {
	// TODO: this will not be a pointer to *base.ClusterContext,
	// as we're moving away from it; it will be its own thing
	Namespace string

	kube kubernetes.Interface
}

func (d Namespace) GetNamespace() string {
	return d.Namespace
}

// This is a helper to get access to the Pods API for this namespace
func (n Namespace) PodInterface() v1.PodInterface {
	return n.kube.CoreV1().Pods(n.Namespace)
}
