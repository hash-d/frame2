package f2skupper1_test

import (
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/disruptor"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1environment"
	"testing"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"gotest.tools/assert"
)

func TestStatusSimple(t *testing.T) {

	r := &frame2.Run{
		T: t,
	}
	defer r.Finalize()

	r.AllowDisruptors([]frame2.Disruptor{
		&disruptor.UpgradeAndFinalize{},
		&disruptor.EdgeOnPrivate{},
		&disruptor.AlternateSkupper{},
		&disruptor.SkipManifestCheck{},
	})

	envSetup := &f2sk1environment.JustSkupperSimple{
		Name:         "status-test",
		AutoTearDown: true,
	}

	setup := frame2.Phase{
		Runner: r,
		Doc:    "Create a skupper installation",
		Setup: []frame2.Step{
			{
				Modify: envSetup},
		},
	}
	assert.Assert(t, setup.Run())

	pub, err := envSetup.Topo.Get(f2k8s.Public, 1)
	assert.Assert(t, err)
	prv, err := envSetup.Topo.Get(f2k8s.Private, 1)
	assert.Assert(t, err)

	basicRetry := frame2.RetryOptions{
		Allow: 120,
	}

	test := frame2.Phase{
		Runner: r,
		Doc:    "Main test phase",
		MainSteps: []frame2.Step{
			{
				Doc: "Execute skupper status on public-1",
				Validator: &f2skupper1.Status{
					Namespace:             pub,
					CheckStatus:           true,
					Enabled:               true,
					CheckConnectionCounts: true,
					TotalConn:             1,
					CheckServiceCount:     true,
					ExposedServices:       0,
					CheckPolicies:         true,
					Policies:              false,
				},
				ValidatorFinal: true,
				ValidatorRetry: basicRetry,
			},
			{
				Doc: "Execute skupper status on private-1",
				Validator: &f2skupper1.Status{
					Namespace:             prv,
					CheckStatus:           true,
					Enabled:               true,
					CheckConnectionCounts: true,
					TotalConn:             1,
					CheckServiceCount:     true,
					ExposedServices:       0,
					CheckPolicies:         true,
					Policies:              false,
				},
				ValidatorFinal: true,
				ValidatorRetry: basicRetry,
			},
			{
				Doc: "Execute skupper -v status on public-1",
				Validator: &f2skupper1.Status{
					Namespace:             pub,
					Verbose:               true,
					CheckStatus:           true,
					Enabled:               true,
					CheckConnectionCounts: true,
					TotalConn:             1,
					DirectConn:            1,
					IndirectConn:          0,
					CheckServiceCount:     true,
					ExposedServices:       0,
					CheckPolicies:         true,
					Policies:              false,
				},
				ValidatorFinal: true,
				ValidatorRetry: basicRetry,
			},
			{
				Doc: "Execute skupper -v status on private-1",
				Validator: &f2skupper1.Status{
					Namespace:             prv,
					Verbose:               true,
					CheckStatus:           true,
					Enabled:               true,
					CheckConnectionCounts: true,
					TotalConn:             1,
					DirectConn:            1,
					IndirectConn:          0,
					CheckServiceCount:     true,
					ExposedServices:       0,
					CheckPolicies:         true,
					Policies:              false,
				},
				ValidatorFinal: true,
				ValidatorRetry: basicRetry,
			},
		},
	}

	assert.Assert(t, test.Execute())

}

func TestStatusHelloWorldN(t *testing.T) {

	r := &frame2.Run{
		T: t,
	}
	defer r.Finalize()

	r.AllowDisruptors([]frame2.Disruptor{
		&disruptor.UpgradeAndFinalize{},
		&disruptor.AlternateSkupper{},
		&disruptor.SkipManifestCheck{},

		// EdgeOnPrivate cannot be used with N topology: edge sites
		// have a single uplink connection; prv2 would the be connected
		// to either pub1 or pub2, and the VAN would be severed (ie,
		// pub1 cannot see prv1)
		// &disruptors.EdgeOnPrivate{},
	})

	envSetup := &f2sk1environment.HelloWorldN{
		Name:          "status-hello-world-n",
		SkupperExpose: true,
		AutoTearDown:  true,
	}

	setup := frame2.Phase{
		Runner: r,
		Doc:    "Create a skupper installation",
		Setup: []frame2.Step{
			{
				Modify: envSetup,
			},
		},
	}
	assert.Assert(t, setup.Run())

	pub1, err := envSetup.Topology.Get(f2k8s.Public, 1)
	assert.Assert(t, err)
	prv1, err := envSetup.Topology.Get(f2k8s.Private, 1)
	assert.Assert(t, err)
	pub2, err := envSetup.Topology.Get(f2k8s.Public, 2)
	assert.Assert(t, err)
	prv2, err := envSetup.Topology.Get(f2k8s.Private, 2)
	assert.Assert(t, err)

	basicRetry := frame2.RetryOptions{
		Timeout:    time.Minute * 2,
		KeepTrying: true,
	}

	test := frame2.Phase{
		Runner: r,
		Doc:    "On a HelloWorld deployed in pub-1 and pub-2 of an N-shaped Van, check the results of skupper status",
		MainSteps: []frame2.Step{
			{
				Doc: "Execute skupper status and status -v on all namespaces",
				Validators: []frame2.Validator{
					&f2skupper1.Status{
						Namespace:             pub1,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						IndirectConn:          2,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
					&f2skupper1.Status{
						Namespace:             pub1,
						Verbose:               true,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						DirectConn:            1,
						IndirectConn:          2,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
					&f2skupper1.Status{
						Namespace:             prv1,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						IndirectConn:          2,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
					&f2skupper1.Status{
						Namespace:             prv1,
						Verbose:               true,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						DirectConn:            1,
						IndirectConn:          2,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
					&f2skupper1.Status{
						Namespace:             pub2,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						IndirectConn:          1,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
					&f2skupper1.Status{
						Namespace:             pub2,
						Verbose:               true,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						DirectConn:            2,
						IndirectConn:          1,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
					&f2skupper1.Status{
						Namespace:             prv2,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						IndirectConn:          1,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
					&f2skupper1.Status{
						Namespace:             prv2,
						Verbose:               true,
						CheckStatus:           true,
						Enabled:               true,
						CheckConnectionCounts: true,
						TotalConn:             3,
						DirectConn:            2,
						IndirectConn:          1,
						CheckServiceCount:     true,
						ExposedServices:       2,
						CheckPolicies:         true,
						Policies:              false,
					},
				},
				ValidatorFinal: true,
				ValidatorRetry: basicRetry,
			},
		},
	}

	assert.Assert(t, test.Execute())

}
