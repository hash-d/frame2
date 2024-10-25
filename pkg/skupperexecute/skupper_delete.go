package skupperexecute

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

type SkupperDelete struct {
	Namespace *f2k8s.Namespace

	Context context.Context
	frame2.DefaultRunDealer
}

// TODO: remove autodebug
// TODO: using old version on post-subfinalizer-hook with UPGRADE_AND_FINALIZE
func (s *SkupperDelete) Execute() error {

	ctx := s.Context
	if s.Context == nil {
		ctx = context.Background()
	}

	phase := frame2.Phase{
		Runner: s.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					F2Namespace: s.Namespace,
					Args:        []string{"delete"},
					Cmd: execute.Cmd{
						Ctx: ctx,
					},
				},
			},
		},
	}
	return phase.Run()

}
