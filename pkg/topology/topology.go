package topology

import (
	"fmt"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/skupperexecute"
	"github.com/skupperproject/skupper/test/utils/base"
)

// Any topology needs to implement this interface.
//
// Its methods allow for components from frame2.deploy or tests to inquire
// about the topology's members, and deal with them without direct access to
// the TopologyMap.
type Basic interface {
	frame2.Executor

	GetTopologyMap() (*TopologyMap, error)

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
	Get(kind f2k8s.ClusterType, number int) (*f2k8s.Namespace, error)

	// This is the same as Get, but it will fail if the number is higher
	// than what the cluster provides.  Use this only if the test requires
	// a specific minimum number of ClusterContexts
	GetStrict(kind f2k8s.ClusterType, number int) (base.ClusterContext, error)

	// Get all clusterContexts of a certain type.  Note this be filtered
	// depending on the topology
	GetAll(kind f2k8s.ClusterType) []*f2k8s.Namespace

	// Same as above, but unfiltered
	GetAllStrict(kind f2k8s.ClusterType) []base.ClusterContext

	// Get a list with all clusterContexts, regardless of type or role
	ListAll() []*f2k8s.Namespace
}

// This represents a Topology which starts with a main branch and a secondary
// branch, connected by a Vertex.
//
// It can be used, for example, to test migrations.   The application and
// Skupper are initially installed on the Left branch, and the test moves it
// to the Right branch.
type TwoBranched interface {
	Basic

	// Same as Basic.Get(), but specifically on the left branch
	GetLeft(kind f2k8s.ClusterType, number int) (*f2k8s.Namespace, error)

	// Same as Basic.Get(), but specifically on the right branch
	GetRight(kind f2k8s.ClusterType, number int) (*f2k8s.Namespace, error)

	// Get the ClusterContext that connects the two branches
	GetVertex() (*f2k8s.Namespace, error)
}

// TopoMap: receives

// A TopologyItem represents a skupper instalation on a namespace.
// The connections are outgoing links to other TopologyItems (or:
// to other Skupper installations)
//
// Once a TopologyItem has been realized by running its TopologyMap,
// the respective ClusterContext will be assigned
type TopologyItem struct {
	Type                  f2k8s.ClusterType
	Connections           []*TopologyItem
	SkipNamespaceCreation bool
	SkipSkupperDeploy     bool
	//SkipApplicationDeploy bool TODO

	Namespace *f2k8s.Namespace

	// An identifier added to the namespace name
	Name string

	// TODO: need to add SkupperInstall configuration for the
	//       topology items, so site-specific configurations
	//       can be done (such as activating the console)
	EnableConsole bool
	// TODO
	EnableFlowCollector bool
}

// TopologyMap receives a list of TopologyItem that describe the topology.
//
// When executed, it creates the required ClusterContexts and returns three items:
//
// - A list of private clusterContexts
// - A list of public  clusterContexts
// - A go map from TopologyItem to ClusterContext
//
// These ClusterContexts do not yet refer to existing namespaces: that, along
// with Skupper installation and creation of the links is done by Topology and
// TopologyConnect.
//
// In general, tests should not use a TopologyMap as an executor.  Instead,
// just define it on a Topology, which will execute it.
//
// clients should keep a reference to a TopologyMap to
// get their output
type TopologyMap struct {
	// This will become the prefix on the name for the namespaces created
	Name         string
	AutoTearDown bool

	TestBase *f2k8s.TestBase

	// Input
	Map []*TopologyItem

	frame2.DefaultRunDealer

	GeneratedMap map[*TopologyItem]*f2k8s.Namespace
}

// Creates the ClusterContext items based on the provided map
//
// The actual namespaces are not yet created on this step.  Give the TopologyMap to a
// TopologyBuild to create them (and everything else)
//
// TODO: Validate: check for duplicates, disconnected items, etc (but allow to skip validation)
func (tm *TopologyMap) Execute() error {
	if tm.Name == "" {
		return fmt.Errorf("TopologyMap configurarion error: no name provided")
	}
	if len(tm.Map) == 0 {
		return fmt.Errorf("TopologyMap configuration error: no topology provided")
	}
	err := TopologyValidator{}.Execute()
	if err != nil {
		return err
	}

	err = f2k8s.ConnectInitial()
	if err != nil {
		return fmt.Errorf("TopologyMap: failed connecting to the clusters: %w", err)
	}

	tm.GeneratedMap = map[*TopologyItem]*f2k8s.Namespace{}

	steps := []frame2.Step{}

	for _, item := range tm.Map {
		log.Printf("item: %+v", item)

		create := &f2k8s.CreateNamespaceTestBase{
			TestBase:     tm.TestBase,
			Kind:         item.Type,
			Id:           item.Name,
			AutoTearDown: tm.AutoTearDown,
		}

		// This closure captures the item and &f2k8s.CreateNamespaceTestBase,
		// and then it saves it on tm.GeneratedMap, so we can do it all in
		// a single phase
		save := func() func() error {
			create := create
			item := item
			return func() error {
				tm.GeneratedMap[item] = &(create.Return)
				item.Namespace = &create.Return
				return nil
			}
		}()

		steps = append(
			steps,
			frame2.Step{
				Modify: create,
			},
			frame2.Step{
				Doc: "Save reference to created namespace",
				Modify: execute.Function{
					Fn: save,
				},
			},
		)
	}

	phase := frame2.Phase{
		Runner:    tm.GetRunner(),
		MainSteps: steps,
	}
	return phase.Run()
}

// TODO: Not yet implemented
type TopologyValidator struct {
	TopologyMap
}

func (tv TopologyValidator) Execute() error {
	log.Printf("TopologyValidator not yet implemented")
	return nil
}

// Based on a Topology, create the VAN:
//
// - Create the namespaces/ClusterContexts
// - Install Skupper
// - Create the links between the nodes.
//
// This ties together Topology, TopologyConnect
// and other items
type TopologyBuild struct {
	// TODO: review this; pointer to interface.
	Topology     *Basic
	AutoTearDown bool

	SkipConnect bool

	// TODO Remove this?
	teardowns []frame2.Executor

	frame2.DefaultRunDealer
	frame2.Log
}

func (t *TopologyBuild) Execute() error {

	steps := frame2.Phase{
		Runner: t.GetRunner(),
		Doc:    "Create the topology",
		MainSteps: []frame2.Step{
			{
				Modify: *t.Topology,
			},
		},
	}
	steps.Run()

	tm, err := (*t.Topology).GetTopologyMap()
	if err != nil {
		return fmt.Errorf("failed to get topologyMap: %w", err)
	}

	// Execute the TopologyMap; create the ClusterContext items
	buildTopologyMap := frame2.Phase{
		Runner: t.Runner,
		Doc:    "Execute the TopologyMap",
		MainSteps: []frame2.Step{
			{
				Modify: tm,
			},
		},
	}
	buildTopologyMap.Run()
	log.Printf("Generated TopologyMap: %+v", tm)

	log.Printf("Creating namespaces and installing Skupper")
	for topoItem, context := range tm.GeneratedMap {

		createAndInstall := frame2.Phase{
			Runner: t.Runner,
			Doc:    "Create namespaces and install Skupper",
			Setup: []frame2.Step{
				{
					/*
							Modify: &f2k8s.CreateNamespaceRaw{
								Namespace:    context,
								AutoTearDown: t.AutoTearDown,
							},
							SkipWhen: topoItem.SkipNamespaceCreation,
						}, {
					*/
					Modify: &skupperexecute.SkupperInstallSimple{
						Namespace:     context,
						EnableConsole: topoItem.EnableConsole,
					},
					SkipWhen: topoItem.SkipNamespaceCreation || topoItem.SkipSkupperDeploy,
				},
			},
		}
		createAndInstall.Run()
	}

	connectSteps := frame2.Phase{
		Runner: t.Runner,
		Setup: []frame2.Step{
			{
				Modify: &TopologyConnect{
					TopologyMap: *tm,
				},
				SkipWhen: t.SkipConnect,
			},
		},
	}
	connectSteps.Run()

	return nil
}

// TODO Perhaps change the frame2.TearDowner interface to return a []frame2.Executor, instead, so a single
// call may return several, and have them run by the Runner?
func (t TopologyBuild) TearDown() frame2.Executor {
	return execute.Function{
		Fn: func() error {
			var ret error
			for _, td := range t.teardowns {
				err := td.Execute()
				if err != nil {
					log.Printf("topology teardown failed: %v", err)
					ret = fmt.Errorf("at least one step of topology teardown failed.  Last error: %w", err)
				}

			}
			return ret
		},
	}
}

type TopologyConnect struct {
	TopologyMap TopologyMap

	frame2.DefaultRunDealer
	// TODO: add some filters and run only one part of the topology
	// 	 (allow for late runs)
	frame2.Log
}

// Assumes that the namespaces are already created, and Skupper installed on all
// namespaces that will create or receive links
func (tc TopologyConnect) Execute() error {

	// TODO change this to something that creates a list of SkupperConnect, then
	// executes it all in a single phase,
	for from, ctx := range tc.TopologyMap.GeneratedMap {
		if from.SkipNamespaceCreation || from.SkipSkupperDeploy {
			continue
		}
		for _, to := range from.Connections {
			pivot := tc.TopologyMap.GeneratedMap[to]
			if to.SkipNamespaceCreation || to.SkipSkupperDeploy {
				continue
			}
			connName := fmt.Sprintf("%v-to-%v", ctx.GetNamespaceName(), pivot.GetNamespaceName())
			phase := frame2.Phase{
				Runner: tc.Runner,
				Doc:    fmt.Sprintf("Creating connection %q", connName),
				MainSteps: []frame2.Step{
					{
						Modify: &skupperexecute.Connect{
							LinkName: connName,
							From:     ctx,
							To:       pivot,
						},
					},
				},
			}
			err := phase.Run()
			if err != nil {
				return fmt.Errorf("TopologyConnect failed: %w", err)
			}
		}
	}

	return nil
}
