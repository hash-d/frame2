package skupperexecute

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

type SkupperUnLink struct {
	Name   string
	From   *f2k8s.Namespace
	To     *f2k8s.Namespace
	Ctx    context.Context
	Runner *frame2.Run
	frame2.Log
}

func (s SkupperUnLink) Execute() error {

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					F2Namespace: s.From,
					Args:        []string{"link", "delete", s.Name},
				},
			},
		},
	}
	return phase.Run()
}
