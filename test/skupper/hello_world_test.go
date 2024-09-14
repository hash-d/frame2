package skupper

import (
	"fmt"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/environment"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/topology"
	"github.com/hash-d/frame2/pkg/topology/topologies"
)

func TestHelloWorld(t *testing.T) {

	testBase := f2k8s.NewTestBase("hello-world")
	runner := frame2.Run{T: t}

	var topologyN topology.Basic
	topologyN = &topologies.N{
		Name:     "hello-n",
		TestBase: testBase,
	}

	prepareTopology := frame2.Phase{
		Runner: &runner,
		Name:   "Prepare the topology",
		Setup: []frame2.Step{
			{
				Modify: topologyN,
			}, {
				Modify: execute.Print{
					Message: fmt.Sprintf("topologyN: %#v", &topologyN),
				},
			},
		},
	}
	prepareTopology.Run()

	deployApp := frame2.Phase{
		Runner: &runner,
		Name:   "Set up HelloWorld",
		Setup: []frame2.Step{
			{
				Modify: &environment.HelloWorld{
					Topology:     &topologyN,
					AutoTearDown: true,
				},
			},
		},
	}
	deployApp.Run()

}
