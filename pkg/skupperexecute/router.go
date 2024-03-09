package skupperexecute

import (
	"context"
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/skupperproject/skupper/test/utils/base"
)

type RouterCheck struct {
	Namespace *base.ClusterContext

	Mode     string
	LogLevel string

	Ctx context.Context

	Return *interface{}

	frame2.DefaultRunDealer
	frame2.Log
}

func (r *RouterCheck) Validate() error {
	phase := frame2.Phase{
		Runner: r.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Doc: fmt.Sprintf("Checking that the router mode is set to %q on skrouterd.conf", r.Mode),
				Validator: &k8svalidate.ConfigMap{
					Namespace:   r.Namespace,
					Name:        "skupper-internal",
					Ctx:         r.Ctx,
					LogContents: true,

					JSON: map[string]f2general.JSON{
						"skrouterd.json": {
							Matchers: []f2general.JSONMatcher{
								{
									Expression: fmt.Sprintf("[?[0] == 'router'] |[].mode | map((&@ == '%v'), @)", r.Mode),
									Exact:      1,
								},
							},
						},
					},
				},
				SkipWhen: r.Mode == "",
			},
		},
	}
	phase.Run()
	return nil

}
