package skupperexecute

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
)

type LinkCreate struct {
	Namespace *base.ClusterContext

	// The token file to be used on the link creation
	File string
	Name string
	Cost string

	frame2.DefaultRunDealer
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
					Args:           args,
					ClusterContext: lc.Namespace,
				},
			},
		},
	}
	return phase.Run()
}
