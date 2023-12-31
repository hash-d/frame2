package skupper

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/topology"
	"github.com/hash-d/frame2/pkg/topology/topologies"
	"github.com/skupperproject/skupper/test/utils/base"
)

func TestTopologyMap(t *testing.T) {
	runner := base.ClusterTestRunnerBase{}

	pub1 := &topology.TopologyItem{
		Type: topology.Public,
	}
	pub2 := &topology.TopologyItem{
		Type: topology.Public,
	}

	prv1 := &topology.TopologyItem{
		Type: topology.Private,
		Connections: []*topology.TopologyItem{
			pub1,
			pub2,
		},
	}
	prv2 := &topology.TopologyItem{
		Type: topology.Private,
		Connections: []*topology.TopologyItem{
			pub2,
		},
	}

	topoMap := []*topology.TopologyItem{
		pub1,
		pub2,
		prv1,
		prv2,
	}

	tm := topology.TopologyMap{
		Name:           "topo",
		TestRunnerBase: &runner,
		Map:            topoMap,
	}
	var custom topology.Basic
	custom = &topologies.Custom{
		TopologyMap: &tm,
	}

	tests := frame2.Phase{
		Name: "TestTopology",
		Setup: []frame2.Step{
			{
				Modify: &tm,
			}, {
				Modify: &topology.TopologyBuild{
					Topology:     &custom,
					AutoTearDown: true,
				},
			},
		},
		MainSteps: []frame2.Step{
			{
				Doc: "Show it to me",
				Modify: execute.Print{
					Data: []interface{}{topoMap},
				},
			},
		},
	}

	tests.RunT(t)
}
