package skupperexecute

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/skupperproject/skupper/test/utils/base"
)

type CliSkupperVersion struct {
	Namespace *base.ClusterContext
	Ctx       context.Context
	frame2.DefaultRunDealer
	execute.SkupperVersionerDefault
}

func (s CliSkupperVersion) Execute() error {

	args := []string{"version"}

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: s.Namespace,
				},
			},
		},
	}

	return phase.Run()
}
