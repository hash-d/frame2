package k8sexecute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/k8sexecute"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateGetDeleteSecret(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	runnerBase := base.ClusterTestRunnerBase{}
	setup := frame2.Phase{
		Runner: &r,
		Setup: []frame2.Step{
			{
				Modify: &execute.BuildClusterContext{
					RunnerBase: &runnerBase,
					Needs: base.ClusterNeeds{
						NamespaceId:    "k8s-secret",
						PublicClusters: 1,
					},
					AutoTearDown: true,
				},
			},
		},
	}
	assert.Assert(t, setup.Run())

	pub1, err := runnerBase.GetPublicContext(1)
	assert.Assert(t, err)

	main := frame2.Phase{
		Runner: &r,
		MainSteps: []frame2.Step{
			{
				Name: "create secret and validate",
				Modify: k8sexecute.SecretCreate{
					Namespace: pub1,
					Secret: &core.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test-secret",
							Namespace: pub1.Namespace,
						},
						Type: core.SecretTypeOpaque,
						Data: map[string][]byte{
							"asdf": []byte(`qwerty`),
							"foo":  []byte("bar"),
						},
					},
				},
				Validator: k8svalidate.SecretGet{
					Namespace: pub1,
					Name:      "test-secret",
					Expect: map[string][]byte{
						"asdf": []byte("qwerty"),
					},
				},
			}, {
				Name: "negative tests",
				Validators: []frame2.Validator{
					k8svalidate.SecretGet{
						Namespace: pub1,
						Name:      "test-secret",
						Expect: map[string][]byte{
							"asdf": []byte("qwerty"),
						},
						ExpectAll: true,
					},
					k8svalidate.SecretGet{
						Namespace: pub1,
						Name:      "test-secret",
						Expect: map[string][]byte{
							"asdf": []byte("bar"),
						},
					},
					k8svalidate.SecretGet{
						Namespace: pub1,
						Name:      "test-secret",
						Expect: map[string][]byte{
							"foo": []byte("qwerty"),
						},
					},
					k8svalidate.SecretGet{
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
				Modify: k8sexecute.SecretDelete{
					Namespace: pub1,
					Name:      "test-secret",
				},
				Validator: k8svalidate.SecretGet{
					Namespace: pub1,
					Name:      "test-secret",
				},
				ExpectError: true,
			},
		},
	}
	assert.Assert(t, main.Run())
}
