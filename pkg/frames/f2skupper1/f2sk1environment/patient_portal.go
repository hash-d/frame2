package f2sk1environment

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1deploy"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology/topologies"
)

// A Patient Portal deployment on pub1 (frontend), prv1 (DB) and prv2 (payment),
// on the default topology
type PatientPortalDefault struct {
	Name         string
	AutoTearDown bool

	// If true, console will be enabled on prv1
	EnableConsole bool

	// Return

	TopoReturn topology.Basic
	frame2.DefaultRunDealer
}

func (p *PatientPortalDefault) Execute() error {

	name := p.Name
	if name == "" {
		name = "patient-portal"
	}

	testBase := f2k8s.NewTestBase(p.Name)

	var topoSimplest topology.Basic
	topoSimplest = &topologies.Simplest{
		Name:             name,
		TestBase:         testBase,
		ConsoleOnPrivate: p.EnableConsole,
	}

	p.TopoReturn = topoSimplest

	execute := &frame2.Phase{
		Runner: p.GetRunner(),
		Doc:    "Default Patient Portal deployment",
		MainSteps: []frame2.Step{
			{
				Modify: &PatientPortal{
					Topology:      &topoSimplest,
					AutoTearDown:  p.AutoTearDown,
					SkupperExpose: true,
				},
			},
		},
	}

	return execute.Run()
}

type PatientPortal struct {
	Topology      *topology.Basic
	AutoTearDown  bool
	SkupperExpose bool

	// If true, console will be enabled on prv1
	EnableConsole bool

	frame2.DefaultRunDealer
}

func (p PatientPortal) Execute() error {
	topo := topology.TopologyBuild{
		Topology:     p.Topology,
		AutoTearDown: p.AutoTearDown,
	}

	execute := frame2.Phase{
		Runner: p.GetRunner(),
		Doc:    "Deploy a Patient Portal environment",
		MainSteps: []frame2.Step{
			{
				Modify: &topo,
			}, {
				Modify: &f2sk1deploy.PatientPortal{
					Topology:      p.Topology,
					SkupperExpose: p.SkupperExpose,
				},
			},
		},
	}
	return execute.Run()
}
