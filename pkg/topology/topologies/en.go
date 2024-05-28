package topologies

import (
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/topology"
)

// Two pub, two private.  Connections always from prv to pub
//
// prv2 has two outgoing links; pub2 has two incoming links
//
//	pub2 pub1
//	 | \  |     ^
//	 |  \ |     |   Connection direction
//	prv1 prv2
//
// # Good for minimal multiple link testing
type N struct {
	Name     string
	TestBase *f2k8s.TestBase

	*contextHolder

	Return *topology.TopologyMap
}

func (n *N) Execute() error {

	pub1 := &topology.TopologyItem{
		Type: f2k8s.Public,
	}
	pub2 := &topology.TopologyItem{
		Type: f2k8s.Public,
	}

	prv1 := &topology.TopologyItem{
		Type: f2k8s.Private,
		Connections: []*topology.TopologyItem{
			pub2,
		},
	}
	prv2 := &topology.TopologyItem{
		Type: f2k8s.Private,
		Connections: []*topology.TopologyItem{
			pub1,
			pub2,
		},
	}

	topoMap := []*topology.TopologyItem{
		pub1,
		pub2,
		prv1,
		prv2,
	}

	n.Return = &topology.TopologyMap{
		Name:     n.Name,
		TestBase: n.TestBase,
		Map:      topoMap,
	}

	n.contextHolder = &contextHolder{TopologyMap: n.Return}

	return nil
}
