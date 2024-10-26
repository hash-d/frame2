package validates

import (
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

// TODO - This should move to a validate package, but it can't be there
// right now because of an import cycle.  I don't want to refactor the
// code before I finish the est
//
// Executes nslookup within a pod, to check whether a name is valid
// within a namespace or cluster
type Nslookup struct {
	Namespace *f2k8s.Namespace

	Name string

	Cmd execute.Cmd

	frame2.Log
	frame2.DefaultRunDealer
}

func (n Nslookup) Validate() error {

	arg := fmt.Sprintf("kubectl --namespace %s exec deploy/dnsutils -- nslookup %q", n.Namespace.GetNamespaceName(), n.Name)

	n.Cmd.Command = arg
	n.Cmd.Shell = true

	phase := frame2.Phase{
		Runner: n.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &n.Cmd,
			},
		},
	}
	return phase.Run()
}
