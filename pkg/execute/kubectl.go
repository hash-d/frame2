package execute

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
)

type Kubectl struct {
	Args []string

	// Secondary way to get the namespace, used only if Namespace is empty
	ClusterContext *base.ClusterContext

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd Cmd

	frame2.Log
	frame2.DefaultRunDealer
}

func (k Kubectl) Execute() error {

	// TODO: add --kubeconfig based on k.ClusterContext

	if k.Cmd.Shell {
		k.Cmd.Command = "kubectl " + k.Cmd.Command
	} else {
		k.Cmd.Command = "kubectl"
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
