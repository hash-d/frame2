package skupper

import (
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology/topologies"
)

func TestTopologyMap(t *testing.T) {
	testBase := f2k8s.NewTestBase("topo-map")

	pub1 := &topology.TopologyItem{
		Type: f2k8s.Public,
	}
	pub2 := &topology.TopologyItem{
		Type: f2k8s.Public,
	}

	prv1 := &topology.TopologyItem{
		Type: f2k8s.Private,
		Connections: []*topology.TopologyItem{
			pub1,
			pub2,
		},
	}
	prv2 := &topology.TopologyItem{
		Type: f2k8s.Private,
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
		Name:         "topo",
		TestBase:     testBase,
		Map:          topoMap,
		AutoTearDown: true,
	}
	var custom topology.Basic
	custom = &topologies.Custom{
		TopologyMap: &tm,
	}

	runner := &frame2.Run{
		T: t,
	}
	tests := frame2.Phase{
		Name:   "TestTopology",
		Runner: runner,
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
				Modify: f2general.Print{
					Data: []interface{}{topoMap},
				},
			},
		},
	}

	tests.Run()
}
