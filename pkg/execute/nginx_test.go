package execute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
)

func TestNginxDeploy(t *testing.T) {
	baseRunner := base.ClusterTestRunnerBase{}

	r := &frame2.Run{
		T: t,
	}
	cc := execute.BuildClusterContext{
		RunnerBase: &baseRunner,
		Needs: base.ClusterNeeds{
			PrivateClusters: 1,
			PublicClusters:  0,
			NamespaceId:     "test-nginx-deploy",
		},
		AutoTearDown: true,
	}
	p1 := frame2.Phase{
		Runner: r,
		Setup: []frame2.Step{
			{
				Modify: &cc,
			},
		},
	}
	p1.Run()

	prv1, err := cc.RunnerBase.GetPrivateContext(1)
	assert.Assert(t, err)

	p2 := frame2.Phase{
		Runner: r,
		MainSteps: []frame2.Step{
			{
				Doc: "Deploy a plain nginx and check the deployment",
				Modify: execute.NginxDeploy{
					Namespace:     prv1,
					ExposeService: true,
				},
				Validators: []frame2.Validator{
					&execute.K8SDeploymentGet{
						Namespace: prv1,
						Name:      "nginx",
					},
					&validate.Curl{
						Namespace:   prv1,
						Url:         "http://nginx:8080",
						Fail400Plus: true,
						DeployCurl:  true,
						Podname:     "curl",
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Allow:  10,
					Ensure: 5,
				},
			},
		},
	}
	p2.Run()

}
