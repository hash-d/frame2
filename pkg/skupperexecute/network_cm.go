package skupperexecute

import (
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/skupperproject/skupper/test/utils/base"
)

// INCOMPLETE TODO
type NetworkStatusConfigMap struct {
	Namespace *base.ClusterContext

	SiteName string

	*frame2.DefaultRunDealer
}

func (n NetworkStatusConfigMap) Validate() error {
	asserter := frame2.Asserter{}

	phase := frame2.Phase{
		Runner: n.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Validator: &k8svalidate.ConfigMap{
					Namespace: n.Namespace,
					Name:      "skupper-network-status",
					JSON: map[string]f2general.JSON{
						"skrouterd.json": f2general.JSON{
							Matchers: []f2general.JSONMatcher{
								{
									Expression: fmt.Sprintf(""),
									Exact:      1,
								},
							},
						},
					},
				},
			},
		},
	}
	asserter.CheckError(phase.Run(), "failed to check skupper-network-status config map")

	return asserter.Error()

}
