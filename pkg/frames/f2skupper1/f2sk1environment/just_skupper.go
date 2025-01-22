package f2sk1environment

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology/topologies"
)

// A Skupper deployment on pub1 (frontend) and prv1 (backend),
type JustSkupperSimple struct {
	Name         string
	AutoTearDown bool
	Console      bool
	SkipConnect  bool

	// Return
	Topo topology.Basic

	frame2.DefaultRunDealer
}

func (j *JustSkupperSimple) Execute() error {

	name := j.Name
	if name == "" {
		name = "just-skupper"
	}

	testBase := f2k8s.NewTestBase(name)

	j.Topo = &topologies.Simplest{
		Name:             name,
		TestBase:         testBase,
		ConsoleOnPublic:  j.Console,
		ConsoleOnPrivate: j.Console,
	}

	execute := frame2.Phase{
		Runner: j.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &JustSkupper{
					Topology:     &j.Topo,
					AutoTearDown: j.AutoTearDown,
					SkipConnect:  j.SkipConnect,
				},
			},
		},
	}

	return execute.Run()
}

// A Skupper deployment on a single namespace
type JustSkupperSingle struct {
	Name         string
	AutoTearDown bool
	Console      bool

	// By default, create a private namespace; this changes it
	Public bool

	// Return
	Topo topology.Basic

	frame2.DefaultRunDealer
}

func (j *JustSkupperSingle) Execute() error {

	name := j.Name
	if name == "" {
		name = "just-skupper"
	}

	testBase := f2k8s.NewTestBase(name)

	kind := f2k8s.Private
	if j.Public {
		kind = f2k8s.Public
	}

	j.Topo = &topologies.Single{
		Name:     name,
		TestBase: testBase,
		Type:     kind,
	}

	execute := frame2.Phase{
		Runner: j.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &JustSkupper{
					Topology:     &j.Topo,
					AutoTearDown: j.AutoTearDown,
				},
			},
		},
	}

	return execute.Run()
}

// A Skupper deployment on pub1 (frontend) and prv1 (backend),
// on an N topology.
//
// Useful for the simplest multiple link testing.
//
// See topology.N for details on the topology.
type JustSkupperN struct {
}

// As the name says, it's just skupper, connected according to the provided
// topology.  For simpler alternatives, see:
//
//   - environment.JustSkupperSimple
//   - environment.JustSkupperN
//   - ...
//   - environment.JustSkupperPlatform is special. It will use
//     whatever topology the current test is asking for, if
//     possible
type JustSkupper struct {
	Topology     *topology.Basic
	AutoTearDown bool
	SkipConnect  bool

	frame2.DefaultRunDealer
	frame2.Log
}

func (j JustSkupper) Execute() error {
	topo := topology.TopologyBuild{
		Topology:     j.Topology,
		AutoTearDown: j.AutoTearDown,
		SkipConnect:  j.SkipConnect,
	}

	execute := frame2.Phase{
		Runner: j.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &topo,
			},
		},
	}
	return execute.Run()
}
