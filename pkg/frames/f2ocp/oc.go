package f2ocp

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

type OcCli struct {
	Args []string

	Namespace *f2k8s.Namespace

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd f2general.Cmd

	frame2.DefaultRunDealer
}

func (k OcCli) Execute() error {

	// TODO: add --kubeconfig based on k.ClusterContext

	if k.Cmd.Shell {
		k.Cmd.Command = "oc " + k.Cmd.Command
	} else {
		k.Cmd.Command = "oc"
	}

	phase := frame2.Phase{
		MainSteps: []frame2.Step{
			{
				Modify: &k.Cmd,
			},
		},
	}

	return phase.Run()
}
