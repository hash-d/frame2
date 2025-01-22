package f2sk1environment

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1deploy"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology/topologies"
)

// A Hello World deployment on pub1 (frontend) and prv1 (backend),
// on the default topology
type HelloWorldDefault struct {
	Name string

	AutoTearDown bool

	frame2.DefaultRunDealer

	topology topology.Basic
}

func (hwd *HelloWorldDefault) Execute() error {

	name := hwd.Name
	if name == "" {
		name = "hello-world"
	}

	testBase := f2k8s.NewTestBase(name)

	hwd.topology = &topologies.Simplest{
		Name:     name,
		TestBase: testBase,
	}

	execute := frame2.Phase{
		Runner: hwd.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &HelloWorld{
					Topology:     &hwd.topology,
					AutoTearDown: hwd.AutoTearDown,
				},
			},
		},
	}

	return execute.Run()
}

func (hwd HelloWorldDefault) GetTopology() topology.Basic {
	return hwd.topology
}

// A Hello World deployment on pub1 (frontend) and prv1 (backend),
// on an N topology.
//
// Useful for the simplest multiple link testing.
//
// See topology.N for details on the topology.
type HelloWorldN struct {
	AutoTearDown  bool
	Name          string
	SkupperExpose bool

	Topology topology.Basic

	frame2.DefaultRunDealer
}

func (h *HelloWorldN) Execute() error {

	testBase := f2k8s.NewTestBase(h.Name)

	h.Topology = &topologies.N{
		Name:     h.Name,
		TestBase: testBase,
	}
	phase := frame2.Phase{
		Runner: h.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Modify: &HelloWorld{
					Topology:      &h.Topology,
					AutoTearDown:  h.AutoTearDown,
					SkupperExpose: h.SkupperExpose,
				},
			},
		},
	}
	return phase.Run()
}

// A Hello World deployment, with configurations.  For simpler
// alternatives, see:
//
//   - environment.HelloWorldSimple
//   - environment.HelloWorldN
//   - ...
//   - environment.HelloWorldPlatform is special. It will use
//     whatever topology the current test is asking for, if
//     possible
//
// To use the auto tearDown, make sure to populate the Runner
type HelloWorld struct {
	Topology      *topology.Basic
	AutoTearDown  bool
	SkupperExpose bool

	frame2.DefaultRunDealer
}

func (hw HelloWorld) Execute() error {
	topo := topology.TopologyBuild{
		Topology:     hw.Topology,
		AutoTearDown: hw.AutoTearDown,
	}

	execute := frame2.Phase{
		Runner: hw.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &topo,
			}, {
				Modify: &f2sk1deploy.HelloWorld{
					Topology:      hw.Topology,
					SkupperExpose: hw.SkupperExpose,
				},
			},
		},
	}
	return execute.Run()
}
