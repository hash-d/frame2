package topologies

import (
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/topology"
)

// A topology in V shape; odd-numbered namespaces go on the
// left branch, even on the right branch.  Except for one
// additional namespace, that can be selected between pub
// and private, and connects both branches.
//
// The constant that pub1 and prv1 are the farthest apart
// possible on the topology is true within the 'left branch'
//
// Similarly, pub2 and prv2 will be the farthest apart on the
// right branch.
//
// The Vertex node will have the console enabled.
type V struct {
	Name     string
	TestBase *f2k8s.TestBase

	EmptyRight bool // If set, do not deploy Skupper or applications on the right branch
	VertexType f2k8s.ClusterType

	*contextHolder
	vertex *topology.TopologyItem

	//For the future
	// VertexConnectionClusterType // whether Vertex should connect to a pub or private cluster
	// Invert right // inverts the selection above for the right branch
	// NumPub, NumPrv. Allow segments

	Return *topology.TopologyMap
}

func (v *V) Execute() error {
	pub1 := &topology.TopologyItem{
		Type: f2k8s.Public,
	}
	prv1 := &topology.TopologyItem{
		Type: f2k8s.Private,
		Connections: []*topology.TopologyItem{
			pub1,
		},
	}
	pub2 := &topology.TopologyItem{
		SkipSkupperDeploy: true,
		Type:              f2k8s.Public,
	}
	prv2 := &topology.TopologyItem{
		SkipSkupperDeploy: true,
		Type:              f2k8s.Private,
		Connections: []*topology.TopologyItem{
			pub2,
		},
	}
	v.vertex = &topology.TopologyItem{
		Type: f2k8s.Public,
		Connections: []*topology.TopologyItem{
			pub1,
			pub2,
		},
		EnableConsole: true,
	}

	topoMap := []*topology.TopologyItem{
		pub1,
		prv1,
		pub2,
		prv2,
		v.vertex,
	}

	v.Return = &topology.TopologyMap{
		Name:     v.Name,
		TestBase: v.TestBase,
		Map:      topoMap,
	}

	v.contextHolder = &contextHolder{TopologyMap: v.Return}

	return nil
}

// Same as Basic.Get(), but specifically on the left branch
func (v *V) GetLeft(kind f2k8s.ClusterType, number int) (*f2k8s.Namespace, error) {
	all := v.contextHolder.GetAll(kind)
	max := len(all)
	if v.vertex.Type == kind {
		max -= 1
	}
	target := number - 1
	index := (target % (max / 2)) * 2
	return all[index], nil

}

// Same as Basic.Get(), but specifically on the right branch
func (v *V) GetRight(kind f2k8s.ClusterType, number int) (*f2k8s.Namespace, error) {
	all := v.contextHolder.GetAll(kind)
	max := len(all)
	if v.vertex.Type == kind {
		max -= 1
	}
	target := number - 1
	index := (target%(max/2))*2 + 1
	return all[index], nil
}

// Get the ClusterContext that connects the two branches
func (v *V) GetVertex() (*f2k8s.Namespace, error) {
	return v.vertex.ClusterContext, nil
}
