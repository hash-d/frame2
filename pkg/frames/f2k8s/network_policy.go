package f2k8s

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkPolicy struct {
	Namespace *Namespace
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
	_, err := n.Namespace.NetworkPolicyInterface().Get(
		ctx,
		n.Name,
		v1.GetOptions{},
	)
	if err != nil {
		return err
	}
	return nil
}
