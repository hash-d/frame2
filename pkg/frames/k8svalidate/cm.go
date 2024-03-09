package k8svalidate

import (
	"context"
	"fmt"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/skupperproject/skupper/test/utils/base"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMap struct {
	Namespace *base.ClusterContext
	Name      string
	Ctx       context.Context

	LogContents bool

	Values map[string]string

	// JSON verification for the keys on the map
	JSON map[string]f2general.JSON

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
	if c.LogContents {
		log.Printf("Contents of CM %q on %q:\n%+v", c.Name, c.Namespace.Namespace, cm.Data)
	}
	for k, v := range c.Values {
		log.Printf("- Checking key %q", k)
		if actual, ok := cm.Data[k]; asserter.Check(ok, "key %q not found on CM %q", k, c.Name) == nil {
			asserter.Check(
				v == actual,
				"values differ for key %q.  expected %q, got %q",
				k, v, actual,
			)
		}
	}
	if c.CMValidator != nil {
		log.Printf("- Running CMValidator")
		asserter.CheckError(c.CMValidator(*cm), "CM Validator failed")
	}

	if len(c.JSON) > 0 {
		JSONvalidators := []frame2.Validator{}
		for k, v := range c.JSON {
			v.Data = cm.Data[k]
			JSONvalidators = append(JSONvalidators, v)
		}
		phase := frame2.Phase{
			Runner: c.GetRunner(),
			Doc:    "Checking JSON contents for ConfigMap",
			MainSteps: []frame2.Step{
				{
					Validators: JSONvalidators,
				},
			},
		}
		asserter.CheckError(phase.Run(), "JSONValidators failed")
	}

	return asserter.Error()
}
