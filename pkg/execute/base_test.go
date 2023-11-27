package execute_test

import (
	"fmt"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/validate"
	"github.com/skupperproject/skupper/test/utils/base"
)

func TestBuildClusterContext(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	clusterBase := &base.ClusterTestRunnerBase{}
	cc := &execute.BuildClusterContext{
		RunnerBase: clusterBase,
		Needs: base.ClusterNeeds{
			NamespaceId:     "test-build-cluster-context",
			PublicClusters:  1,
			PrivateClusters: 1,
		},
		AutoTearDown: true,
	}
	p1 := frame2.Phase{
		Runner: &r,
		Setup: []frame2.Step{
			{
				Modify: cc,
			},
		},
	}
	p1.Run()

	prv1, err := cc.RunnerBase.GetPrivateContext(1)
	if err != nil {
		t.Fatalf("Failed to get prv1")
	}

	p2 := frame2.Phase{
		Runner: &r,
		MainSteps: []frame2.Step{
			{
				Validator: &validate.Executor{
					Executor: &execute.Kubectl{
						Args: []string{
							"get", "ns",
						},
						ClusterContext: prv1,
						Cmd: execute.Cmd{
							ForceOutput: true,
							Command:     fmt.Sprintf("get ns %s", prv1.Namespace),
							Shell:       true,
						},
					},
				},
			},
		},
	}
	p2.Run()

}
