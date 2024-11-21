package execute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"gotest.tools/assert"
)

func TestNginxDeploy(t *testing.T) {

	r := &frame2.Run{
		T: t,
	}
	assert.Assert(t, f2k8s.ConnectInitial())
	testBase := f2k8s.NewTestBase("nginx")
	ns := &f2k8s.CreateNamespaceTestBase{
		TestBase:     testBase,
		AutoTearDown: true,
		Kind:         f2k8s.Public,
	}
	setup := frame2.Phase{
		Runner: r,
		MainSteps: []frame2.Step{
			{
				Modify: ns,
			},
		},
	}
	assert.Assert(t, setup.Run())

	prv1 := &ns.Return

	p2 := frame2.Phase{
		Runner: r,
		MainSteps: []frame2.Step{
			{
				Doc: "Deploy a plain nginx and check the deployment",
				Modify: f2k8s.NginxDeploy{
					Namespace:     prv1,
					ExposeService: true,
				},
				Validators: []frame2.Validator{
					&f2k8s.K8SDeploymentGet{
						Namespace: prv1,
						Name:      "nginx",
					},
					&f2k8s.Curl{
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
