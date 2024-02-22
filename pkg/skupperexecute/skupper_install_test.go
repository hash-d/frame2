package skupperexecute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/disruptors"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/hash-d/frame2/pkg/skupperexecute"
	"github.com/hash-d/frame2/pkg/topology"
	"github.com/hash-d/frame2/pkg/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

func TestSkupperInstallEffects(t *testing.T) {

	baseRunner := &base.ClusterTestRunnerBase{}
	runner := &frame2.Run{
		T: t,
	}
	runner.AllowDisruptors([]frame2.Disruptor{
		&disruptors.UpgradeAndFinalize{},
		//
	})

	var topo topology.Basic

	topo = &topologies.Single{
		Name:              "skupper-install-effects",
		TestRunnerBase:    baseRunner,
		Type:              topology.Public,
		SkipSkupperDeploy: true,
	}
	build := topology.TopologyBuild{
		Topology:     &topo,
		AutoTearDown: true,
	}

	setup := frame2.Phase{
		Runner: runner,
		Setup: []frame2.Step{
			{
				Modify: &build,
			},
		},
	}
	assert.Assert(t, setup.Run())

	ns, err := topo.Get(topology.Public, 1)
	assert.Assert(t, err)

	phase := frame2.Phase{
		Doc:    "Effects",
		Runner: runner,
		MainSteps: []frame2.Step{
			{
				Substep: &frame2.Effects[skupperexecute.CliSkupperInstall]{
					ExecutionProfile: frame2.BOTH,
					BaseFrame: skupperexecute.CliSkupperInstall{
						Namespace: ns,
					},
					TearDown: []frame2.Step{
						{
							Modify: &skupperexecute.SkupperDelete{
								Namespace: ns,
							},
						},
					},
					Effects: map[string]frame2.CauseEffect[skupperexecute.CliSkupperInstall]{
						"annotations": {
							Patch: skupperexecute.CliSkupperInstall{
								Annotations: []string{
									"skupper.io/qe-test=true",
									"skupper.io/qe-test-name=annotations",
								},
							},
							Validators: []frame2.Validator{
								&k8svalidate.Pods{
									Namespace: ns,
									Labels: map[string]string{
										"app.kubernetes.io/part-of": "skupper",
									},
									OtherAnnotations: map[string]string{
										"skupper.io/qe-test":      "true",
										"skupper.io/qe-test-name": "annotations",
									},
								},
							},
						},
						"test2": {
							Patch: skupperexecute.CliSkupperInstall{
								EnableConsole:       true,
								EnableFlowCollector: true,
							},
						},
						//
					},
				},
			},
		},
	}

	assert.Assert(t, phase.Run())

}
