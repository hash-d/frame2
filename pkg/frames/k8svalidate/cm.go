package k8svalidate

import (
	"context"
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMap struct {
	Namespace *base.ClusterContext
	Name      string
	Ctx       context.Context

	Values map[string]string

	// Function to run against the CM, to validate.  Provides a way
	// to execute more complex validations not available above, inline
	CMValidator func(corev1.ConfigMap) error

	Result *[]corev1.ConfigMap

	frame2.Log
	frame2.DefaultRunDealer
}

func (c *ConfigMap) Validate() error {
	ctx := frame2.ContextOrDefault(c.Ctx)
	asserter := frame2.Asserter{}
	cm, err := c.Namespace.VanClient.KubeClient.CoreV1().ConfigMaps(c.Namespace.Namespace).Get(
		ctx,
		c.Name,
		v1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed retrieving cm %q: %v", c.Name, err)
	}
	for k, v := range c.Values {
		if actual, ok := cm.Data[k]; asserter.Check(ok, "key %q not found on CM %q", k, c.Name) != nil {
			asserter.Check(
				v == actual,
				"values differ for key %q.  expected %q, got %q",
				k, v, actual,
			)
		}
	}

	return asserter.Error()
}
