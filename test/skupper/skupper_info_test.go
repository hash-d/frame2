package skupper

import (
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1environment"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"gotest.tools/assert"
)

func TestSkupperInfo(t *testing.T) {

	run := &frame2.Run{
		T: t,
	}

	installCurrent := f2sk1environment.JustSkupperSimple{
		Name:    "skupper-info",
		Console: true,
		//AutoTearDown: true,
	}
	//installOld := environment.JustSkupperDefault{ }

	SetupPhase := frame2.Phase{
		Runner: run,
		Doc:    "Setup a namespace with Skupper, no other deployments",
		Setup: []frame2.Step{
			{
				Modify: &installCurrent,
			},
		},
	}
	assert.Assert(t, SetupPhase.Run())

	namespace, err := installCurrent.Topo.Get(f2k8s.Public, 1)
	assert.Assert(t, err)

	getInfoCurrent := f2skupper1.SkupperInfo{
		Namespace: namespace,
	}

	infoPhase := frame2.Phase{
		Runner: run,
		Doc:    "Get Skupper information, to compare to manifest.json",
		MainSteps: []frame2.Step{
			{
				Validator: &getInfoCurrent,
			},
		},
	}
	assert.Assert(t, infoPhase.Run())

	testPhase := frame2.Phase{
		Runner: run,
		Doc:    "Compare manifest.json to Skupper info acquired priorly",
		MainSteps: []frame2.Step{
			{
				Modify: f2general.Function{
					Fn: func() error {
						if getInfoCurrent.Result.HasRouter {
							return nil
						}
						return fmt.Errorf("The namespace %q has no skupper-router, so we can't consider it for manifest check", namespace.GetNamespaceName())
					},
				},
				Validator: &f2skupper1.SkupperManifest{
					Expected: getInfoCurrent.Result.Images,
				},
			},
		},
	}
	assert.Assert(t, testPhase.Run())

}
