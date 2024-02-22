package topologies

import (
	"github.com/hash-d/frame2/pkg/topology"
	"github.com/skupperproject/skupper/test/utils/base"
)

// This barely counts as a Skupper topology: it's a single namespace.
type Single struct {
	Name           string
	TestRunnerBase *base.ClusterTestRunnerBase

	SkipSkupperDeploy bool

	Type topology.ClusterType

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
		Name:           s.Name,
		TestRunnerBase: s.TestRunnerBase,
		Map:            topoMap,
	}

	s.contextHolder = &contextHolder{TopologyMap: s.Return}

	return nil
}
