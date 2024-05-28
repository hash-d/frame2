package topologies

import (
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/topology"
)

// It's the simplest Skupper topology you can get: prv1 connects
// to pub1.  That's it.
type Simplest struct {
	Name     string
	TestBase *f2k8s.TestBase

	ConsoleOnPublic  bool
	ConsoleOnPrivate bool

	// Identifier of the public namespace; defaults to frontend if
	// undefined
	PubName string

	// Identifier of the private namespace; defaults to backend if
	// undefined
	PrvName string

	// Add on
	*contextHolder

	Return *topology.TopologyMap
}

func (s *Simplest) Execute() error {

	pubname := s.PubName
	prvname := s.PrvName
	if pubname == "" {
		pubname = "frontend"
	}
	if prvname == "" {
		prvname = "backend"
	}

	pub1 := &topology.TopologyItem{
		Name:                pubname,
		Type:                f2k8s.Public,
		EnableConsole:       s.ConsoleOnPublic,
		EnableFlowCollector: s.ConsoleOnPublic,
	}
	prv1 := &topology.TopologyItem{
		Name:                prvname,
		Type:                f2k8s.Private,
		EnableConsole:       s.ConsoleOnPrivate,
		EnableFlowCollector: s.ConsoleOnPrivate,
		Connections: []*topology.TopologyItem{
			pub1,
		},
	}
	topoMap := []*topology.TopologyItem{
		pub1,
		prv1,
	}

	s.Return = &topology.TopologyMap{
		Name:     s.Name,
		TestBase: s.TestBase,
		Map:      topoMap,
	}

	s.contextHolder = &contextHolder{TopologyMap: s.Return}

	return nil
}
