package skupperexecute_test

import (
	"testing"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/disruptors"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/hash-d/frame2/pkg/skupperexecute"
	"github.com/hash-d/frame2/pkg/subrunner"
	"github.com/hash-d/frame2/pkg/topology"
	"github.com/hash-d/frame2/pkg/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestSkupperInstallEffects(t *testing.T) {

	baseRunner := &base.ClusterTestRunnerBase{}
	runner := &frame2.Run{
		T: t,
	}
	runner.AllowDisruptors([]frame2.Disruptor{
		&disruptors.UpgradeAndFinalize{},
		&disruptors.ConsoleAuth{},
		&disruptors.SkipManifestCheck{},
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
		Doc:    "Create a single, skupper-less namespace",
		Setup: []frame2.Step{
			{
				Modify: &build,
			},
		},
	}
	assert.Assert(t, setup.Run())

	ns, err := topo.Get(topology.Public, 1)
	assert.Assert(t, err)

	basicWait := frame2.RetryOptions{
		Timeout:    time.Minute * 2,
		KeepTrying: true,
	}

	phase := subrunner.Effects[skupperexecute.CliSkupperInstall, *skupperexecute.CliSkupperInstall]{
		ExecutionProfile: subrunner.INDIVIDUAL,
		BaseFrame: &skupperexecute.CliSkupperInstall{
			Namespace: ns,
		},
		TearDown: []frame2.Step{
			{
				Modify: &skupperexecute.SkupperDelete{
					Namespace: ns,
				},
				Validator: &k8svalidate.Pods{
					Namespace: ns,
					Labels: map[string]string{
						"app.kubernetes.io/part-of": "skupper",
					},
					ExpectNone: true,
				},
				ValidatorRetry: frame2.RetryOptions{
					KeepTrying: true,
					Timeout:    time.Minute * 2,
				},
			},
		},
		Combos: map[string][]string{
			// Ensures all deployable resources are present, and check annotations
			"annotations-full": []string{"console", "annotations"},
		},
		// TODO: move all validators below to individual frames.
		// - Something like validateskupper.Console{}
		// Those frames should
		// - Be version-aware
		// - Allow to pick what type of testing to run (filter)
		//   - K8S resource verification
		//   - Application effect (eg curl)
		//   - or: white box vs black box
		Effects: map[string]subrunner.CauseEffect[skupperexecute.CliSkupperInstall]{
			"defaults": {
				// Do not do combos with this one, as its validations might conflict
				// with other CauseEffect items
				Doc: "Confirm a plain skupper install is successful and with expected default values",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       false,
					EnableFlowCollector: false,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.Pods{
						Namespace: ns,
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "skupper",
						},
						ExpectExactly:         2,
						NegativeContainerList: []string{"flow-collector"},
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values: map[string]string{
							"console":                "false",
							"console-authentication": "internal",
							"flow-collector":         "false",
						},
					},
				},
			},
			"annotations": {
				Doc: "Annotations must be present on all elements created by Skupper",
				Patch: skupperexecute.CliSkupperInstall{
					Annotations: []string{
						"skupper.io/qe-test=true",
						"skupper.io/qe-test-name=annotations",
					},
				},
				ValidatorsRetry: basicWait,
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
					// TODO: others, too?  Services, etc
				},
			},
			"flow-collector": {
				Doc: "Check flow collector without the console",
				Patch: skupperexecute.CliSkupperInstall{
					EnableFlowCollector: true,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.Pods{
						Namespace:     ns,
						Labels:        map[string]string{"skupper.io/component": "service-controller"},
						ContainerList: []string{"flow-collector"},
						ExpectExactly: 1,
					},
					&k8svalidate.Pods{
						Namespace: ns,
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "skupper",
						},
						ExpectMin: 3,
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values: map[string]string{
							"console":                "false",
							"console-authentication": "internal",
							"flow-collector":         "true",
						},
					},
				},
			},
			"console": {
				Doc: "Check a basic console",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.Pods{
						Namespace:     ns,
						Labels:        map[string]string{"skupper.io/component": "service-controller"},
						ExpectPhase:   corev1.PodRunning,
						ContainerList: []string{"flow-collector"},
						ExpectExactly: 1,
					},
					&k8svalidate.Pods{
						Namespace: ns,
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "skupper",
						},
						ExpectMin: 3,
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values: map[string]string{
							"console":                "true",
							"console-authentication": "internal",
							"flow-collector":         "true",
						},
					},
				},
			},
			//
		},
	}.GetPhase(runner)

	assert.Assert(t, phase.Run())

}
