package f2skupper1_test

import (
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/disruptor"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1environment"
	"testing"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"gotest.tools/assert"
)

func TestConnectSimple(t *testing.T) {
	runner := frame2.Run{
		T: t,
	}
	runner.AllowDisruptors([]frame2.Disruptor{
		&disruptor.UpgradeAndFinalize{},
		&disruptor.EdgeOnPrivate{},
		&disruptor.AlternateSkupper{},
		&disruptor.SkipManifestCheck{},
	})
	phase := frame2.Phase{
		Runner: &runner,
		MainSteps: []frame2.Step{
			{
				Name:      "basic",
				Doc:       "A simple connection, without any arguments",
				Validator: &ConnectTester{},
			}, {
				Name: "cost",
				Doc:  "Setting a specific cost for the connection",
				Validator: &ConnectTester{
					Cost: "3",
				},
			}, {
				Name: "expired",
				Doc:  "A token that will expire immediately; the connection should fail",
				Validator: &ConnectTester{
					Expiry:      "1ns",
					ExpectError: true,
				},
			},
		},
	}
	assert.Assert(t, phase.Run())

}

// Create a public and a private namespace, disconnected, then
// use skupperexecute.Connnect to connec them, as configured,
// and validate the link creation
type ConnectTester struct {
	SecretName string
	Expiry     string
	Password   string
	TokenType  string
	Uses       string

	LinkName string
	Cost     string

	ExpectError bool

	frame2.DefaultRunDealer
	frame2.Log
}

func (ct ConnectTester) Validate() error {

	envSetup := &f2sk1environment.JustSkupperSimple{
		Name:         "connect-test",
		AutoTearDown: true,
		// The connection will be done below, by the tests
		SkipConnect: true,
	}

	// We could have a single namespace and create the all the links on it.
	// However, we'd not be able to test disruptors such as the upgrade with it
	setup := frame2.Phase{
		Runner: ct.Runner,
		Doc:    "Create two disconnected skupper installations",
		Setup: []frame2.Step{
			{
				Modify: envSetup},
		},
	}
	if err := setup.Run(); err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	// We do not care whether pub or private; just pick the first one
	// (there should be only one)
	pub, err := envSetup.Topo.Get(f2k8s.Public, 1)
	if err != nil {
		return fmt.Errorf("failed to get pub ns: %w", err)
	}
	prv, err := envSetup.Topo.Get(f2k8s.Private, 1)
	if err != nil {
		return fmt.Errorf("failed to get prv ns: %w", err)
	}

	basicWait := frame2.RetryOptions{
		Timeout:    time.Minute * 4,
		Ignore:     10,
		KeepTrying: true,
	}

	linkName := ct.LinkName
	if linkName == "" {
		linkName = "test-link"
	}
	cost := ct.Cost
	if cost == "" {
		cost = "1"
	}

	phase := frame2.Phase{
		Runner: ct.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &f2skupper1.Connect{
					From:       prv,
					To:         pub,
					SecretName: ct.SecretName,
					Expiry:     ct.Expiry,
					Password:   ct.Password,
					TokenType:  ct.TokenType,
					Uses:       ct.Uses,
					LinkName:   linkName,
					Cost:       ct.Cost,
				},
				Validators: []frame2.Validator{
					&f2skupper1.OutgoingLinkCheck{
						Namespace: prv,
						Name:      linkName,
						Cost:      cost,
					},
					&f2skupper1.Status{
						Namespace:             pub,
						CheckConnectionCounts: true,
						TotalConn:             1,
					},
					&f2skupper1.Status{
						Namespace:             prv,
						CheckConnectionCounts: true,
						TotalConn:             1,
					},
				},
				ValidatorRetry:    basicWait,
				ValidatorSubFinal: true,
				ExpectError:       ct.ExpectError,
			},
		},
	}
	return phase.Run()

}
