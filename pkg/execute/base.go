package execute

import (
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
)

type BuildClusterContext struct {
	RunnerBase   *base.ClusterTestRunnerBase
	Needs        base.ClusterNeeds
	AutoTearDown bool

	frame2.Log
	frame2.DefaultRunDealer
}

func (b BuildClusterContext) Execute() error {
	if b.RunnerBase == nil {
		b.RunnerBase = &base.ClusterTestRunnerBase{}
	}
	err := b.RunnerBase.Validate(b.Needs)
	if err != nil {
		return fmt.Errorf("failed to validate needs: %w", err)
	}
	contexts, err := b.RunnerBase.Build(b.Needs, nil)
	if err != nil {
		return fmt.Errorf("failed to build RunnerBase: %w", err)
	}

	for _, c := range contexts {
		p := frame2.Phase{
			Runner: b.GetRunner(),
			Setup: []frame2.Step{
				{
					Modify: TestRunnerCreateNamespace{
						Namespace:    c,
						AutoTearDown: b.AutoTearDown,
					},
				},
			},
		}
		err = p.Run()
		if err != nil {
			return fmt.Errorf("failed to create clusterContext: %w", err)
		}
	}

	return nil
}
