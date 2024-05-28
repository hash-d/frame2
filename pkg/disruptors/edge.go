package disruptors

import (
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/skupperexecute"
)

// Every skupper site on a private cluster will be created with
// RouterMode = edge
//
// TODO: perhaps move this to a SkupperInstall disruptors file, instead?
type EdgeOnPrivate struct {
}

func (n EdgeOnPrivate) DisruptorEnvValue() string {
	return "EDGE_ON_PRIVATE"
}

// TODO: when moving to SkuperOps, will these things be done on the Ops or on
// the UI frames?
func (n *EdgeOnPrivate) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if mod, ok := step.Modify.(*skupperexecute.CliSkupperInstall); ok {
		if mod.Namespace.GetKind() == f2k8s.Private {
			log.Printf(
				"EdgeOnPrivate disruptor updating installation on %q to use edge mode",
				mod.Namespace.GetNamespaceName(),
			)
			mod.RouterMode = "edge"
		}
	}
}
