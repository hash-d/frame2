package k8svalidate

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkPolicy struct {
	Namespace *base.ClusterContext
	Name      string
	Ctx       context.Context

	Values map[string]string

	// Function to run against the Network Policy, to validate.  Provides a way
	// to execute more complex validations not available above, inline
	NMValidator func(corev1.ConfigMap) error

	Result *[]corev1.ConfigMap

	frame2.Log
	frame2.DefaultRunDealer
}

func (n *NetworkPolicy) Validate() error {
	ctx := frame2.ContextOrDefault(n.Ctx)
	//	asserter := frame2.Asserter{}
	_, err := n.Namespace.VanClient.KubeClient.NetworkingV1().NetworkPolicies(n.Namespace.Namespace).Get(
		ctx,
		n.Name,
		v1.GetOptions{},
	)
	if err != nil {
		return err
	}
	return nil
}
