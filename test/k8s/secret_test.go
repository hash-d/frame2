package k8s_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology/topologies"
	"gotest.tools/assert"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateGetDeleteSecret(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	testBase := f2k8s.NewTestBase("k8s-secret")
	var topo topology.Basic
	topo = &topologies.Single{
		Name:              "k8s-secret",
		TestBase:          testBase,
		SkipSkupperDeploy: true,
	}
	setup := frame2.Phase{
		Runner: &r,
		Setup: []frame2.Step{
			{
				Modify: topo,
			},
		},
	}
	assert.Assert(t, setup.Run())
	build := frame2.Phase{
		Runner: &r,
		MainSteps: []frame2.Step{
			{
				Modify: &topology.TopologyBuild{
					Topology:     &topo,
					AutoTearDown: true,
				},
			},
		},
	}
	assert.Assert(t, build.Run())

	pub1, err := topo.Get(f2k8s.Public, 1)
	assert.Assert(t, err)

	main := frame2.Phase{
		Runner: &r,
		MainSteps: []frame2.Step{
			{
				Name: "create secret and validate",
				Modify: f2k8s.SecretCreate{
					Namespace: pub1,
					Secret: &core.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test-secret",
							Namespace: pub1.GetNamespaceName(),
						},
						Type: core.SecretTypeOpaque,
						Data: map[string][]byte{
							"asdf": []byte(`qwerty`),
							"foo":  []byte("bar"),
						},
					},
				},
				Validator: f2k8s.SecretGet{
					Namespace: pub1,
					Name:      "test-secret",
					Expect: map[string][]byte{
						"asdf": []byte("qwerty"),
					},
				},
			}, {
				Name: "negative tests",
				Validators: []frame2.Validator{
					f2k8s.SecretGet{
						Namespace: pub1,
						Name:      "test-secret",
						Expect: map[string][]byte{
							"asdf": []byte("qwerty"),
						},
						ExpectAll: true,
					},
					f2k8s.SecretGet{
						Namespace: pub1,
						Name:      "test-secret",
						Expect: map[string][]byte{
							"asdf": []byte("bar"),
						},
					},
					f2k8s.SecretGet{
						Namespace: pub1,
						Name:      "test-secret",
						Expect: map[string][]byte{
							"foo": []byte("qwerty"),
						},
					},
					f2k8s.SecretGet{
						Namespace: pub1,
						Name:      "test-secret",
						Expect: map[string][]byte{
							"foobar": []byte("this should not exist"),
						},
					},
				},
				ExpectError: true,
			}, {
				Name: "delete-secret",
				Modify: f2k8s.SecretDelete{
					Namespace: pub1,
					Name:      "test-secret",
				},
				Validator: f2k8s.SecretGet{
					Namespace: pub1,
					Name:      "test-secret",
				},
				ExpectError: true,
			},
		},
	}
	assert.Assert(t, main.Run())
}
