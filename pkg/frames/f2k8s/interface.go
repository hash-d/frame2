package f2k8s

import (
	configv1 "github.com/openshift/api/config/v1"
)

// The type of cluster:
//
// - Public
// - Private
// - DMZ
//
// Currently, only the first two are implemented
type ClusterType string

const (
	Public  ClusterType = "pub"
	Private ClusterType = "prv"
	DMZ     ClusterType = "dmz"
)

type Namespacer interface {
	GetNamespace() string
	Kube() configv1.KubeClientConfig
}
