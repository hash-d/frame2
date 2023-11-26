package execute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
)

func TestNginxDeploy(t *testing.T) {
	//baseRunner := base.ClusterTestRunnerBase{}

	r := &frame2.Run{
		T: t,
	}
	p := frame2.Phase{
		Runner: r,
		Setup:  []frame2.Step{
			//			execute.K
		},
	}
	p.Run()
}
