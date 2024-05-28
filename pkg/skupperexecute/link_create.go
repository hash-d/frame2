package skupperexecute

import (
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/skupperproject/skupper/test/utils/base"
)

type LinkCreate struct {
	Namespace *f2k8s.Namespace

	// The token file to be used on the link creation
	File string
	Name string
	Cost string

	frame2.DefaultRunDealer
	execute.SkupperVersionerDefault
}

// TODO: replace this by f2k8s.Namespace
func (l LinkCreate) GetNamespace() string {
	return l.Namespace.GetNamespaceName()
}

func (lc *LinkCreate) Execute() error {
	args := []string{"link", "create"}

	if lc.Name != "" {
		args = append(args, "--name", lc.Name)
	}

	if lc.Cost != "" {
		args = append(args, "--cost", lc.Cost)
	}

	if lc.File != "" {
		args = append(args, lc.File)
	}

	phase := frame2.Phase{
		Runner: lc.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:        args,
					F2Namespace: lc.Namespace,
				},
			},
		},
	}
	return phase.Run()
}

type OutgoingLinkCheck struct {
	Namespace *base.ClusterContext
	Name      string
	Cost      string

	frame2.DefaultRunDealer
	frame2.Log
}

func (o *OutgoingLinkCheck) Validate() error {
	phase := frame2.Phase{
		Runner: o.Runner,
		MainSteps: []frame2.Step{
			{
				Validators: []frame2.Validator{
					&k8svalidate.SecretGet{
						Namespace: o.Namespace,
						Name:      o.Name,
						Annotations: f2general.MapCheck{
							Values: map[string]string{"skupper.io/cost": o.Cost},
						},
						Labels: f2general.MapCheck{
							Values: map[string]string{"skupper.io/type": "connection-token"},
						},
					},
					&k8svalidate.ConfigMap{
						Namespace: o.Namespace,
						Name:      "skupper-internal",
						JSON: map[string]f2general.JSON{
							"skrouterd.json": f2general.JSON{
								Matchers: []f2general.JSONMatcher{
									{
										Expression: fmt.Sprintf(
											"[? [0] == 'connector' && [1].name == '%s' ] | [].cost | map((&@ == `%s`), @)",
											o.Name, o.Cost,
										),
										Exact: 1,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return phase.Run()
}
