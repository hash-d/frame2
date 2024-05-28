package topologies

import (
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/topology"
)

// This barely counts as a Skupper topology: it's a single namespace.
type Single struct {
	Name     string
	TestBase *f2k8s.TestBase

	SkipSkupperDeploy bool

	Type f2k8s.ClusterType

	// Add on
	*contextHolder

	Return *topology.TopologyMap
}

func (s *Single) Execute() error {

	ns := &topology.TopologyItem{
		Type:              s.Type,
		SkipSkupperDeploy: s.SkipSkupperDeploy,
	}
	topoMap := []*topology.TopologyItem{
		ns,
	}

	s.Return = &topology.TopologyMap{
		Name:     s.Name,
		TestBase: s.TopologyMap.TestBase,
		Map:      topoMap,
	}

	s.contextHolder = &contextHolder{TopologyMap: s.Return}

	return nil
}
