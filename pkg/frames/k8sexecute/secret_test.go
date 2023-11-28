package k8sexecute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/k8sexecute"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

func TestCreateGetDeleteSecret(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	runnerBase := base.ClusterTestRunnerBase{}
	setup := frame2.Phase{
		Runner: &r,
		Setup: []frame2.Step{
			{
				Modify: &execute.BuildClusterContext{
					RunnerBase: &runnerBase,
					Needs: base.ClusterNeeds{
						NamespaceId:    "k8s-secret",
						PublicClusters: 1,
					},
					AutoTearDown: true,
				},
			},
		},
	}
	assert.Assert(t, setup.Run())
	main := frame2.Phase{
		Runner: &r,
		MainSteps: []frame2.Step{
			{
				Modify:    k8sexecute.SecretCreate{},
				Validator: k8svalidate.SecretGet{},
			}, {
				Modify:      k8sexecute.SecretDelete{},
				Validator:   k8svalidate.SecretGet{},
				ExpectError: true,
			},
		},
	}
	assert.Assert(t, main.Run())
}
