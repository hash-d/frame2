package topologies

import (
	"fmt"

	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/topology"
	"github.com/skupperproject/skupper/test/utils/base"
)

// This is an add-on for the topologies in this package.  When embedded into
// a topology struct, it will provide the functions that implement
// topology.Basic.
//
// Take note, however, to have its TopologyMap match the topology's (when
// setting it up, and in case that value is changed on the topology for
// any reason)
type contextHolder struct {
	TopologyMap *topology.TopologyMap
}

//func (c *ContextHolder) Execute() error {
//	panic("not implemented") // TODO: Implement
//}

func (c *contextHolder) GetTopologyMap() (*topology.TopologyMap, error) {
	if c.TopologyMap == nil {
		return nil, fmt.Errorf("ContextHolder: no TopologyMap defined")
	}
	return c.TopologyMap, nil
}

// Return a ClusterContext of the given type and number.
//
// Negative numbers count from the end.  So, Get for -1 will return
// the clusterContext with the greatest number of that type.
//
// Attention that for some types of topologies (suc as TwoBranched)
// only part of the clustercontexts may be considered (for example,
// only the left branch)
//
// The number divided against number of contexts of that type on
// the topology, and the remainder will be used.  That allows for
// tests that usually run with several namespace to run also with
// a smaller number.  For example, on a cluster with 4 private
// cluster, a request for number 6 will actually return number 2
func (c *contextHolder) Get(kind f2k8s.ClusterType, number int) (*f2k8s.Namespace, error) {
	if c.TopologyMap == nil {
		return nil, fmt.Errorf("topology has not yet been run")
	}
	kindList := c.GetAll(kind)
	// TODO: implement mod logic, implement negative logic
	// TODO: this should all probably move to a add-on struct
	if len(kindList) == 0 {
		return nil, fmt.Errorf("no clusterContext of type %v on the topology", kind)
	}
	var target int
	if number < 0 {
		target = len(kindList) + (number-1)%len(kindList) - 1
	} else {
		target = (number - 1) % len(kindList)
	}
	return kindList[target], nil
}

// This is the same as Get, but it will fail if the number is higher
// than what the cluster provides.  Use this only if the test requires
// a specific minimum number of ClusterContexts
func (c *contextHolder) GetStrict(kind f2k8s.ClusterType, number int) (base.ClusterContext, error) {
	panic("not implemented") // TODO: Implement
}

// Get all clusterContexts of a certain type.  Note this be filtered
// depending on the topology
func (c *contextHolder) GetAll(kind f2k8s.ClusterType) []*f2k8s.Namespace {
	switch kind {
	case f2k8s.Public, f2k8s.Private:
		return c.TopologyMap.TestBase.GetDomainNamespaces(kind)
	}
	panic("Only public and private implemented")

}

// Same as above, but unfiltered
func (c *contextHolder) GetAllStrict(kind f2k8s.ClusterType) []base.ClusterContext {
	panic("not implemented") // TODO: Implement
}

// Get a list with all clusterContexts, regardless of type or role
func (c *contextHolder) ListAll() []*f2k8s.Namespace {
	ret := []*f2k8s.Namespace{}
	ret = append(ret, c.TopologyMap.TestBase.GetDomainNamespaces(f2k8s.Public)...)
	ret = append(ret, c.TopologyMap.TestBase.GetDomainNamespaces(f2k8s.Private)...)
	return ret
}
