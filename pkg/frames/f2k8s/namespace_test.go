package f2k8s

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"gotest.tools/assert"
)

func TestCreateRaw(t *testing.T) {
	r := &frame2.Run{
		T: t,
	}

	err := ConnectInitial()
	assert.Assert(t, err)
	cluster := clusters[0]

	phase := frame2.Phase{
		Runner: r,
		Setup: []frame2.Step{
			{
				Modify: &NamespaceCreateRaw{
					Cluster:      cluster,
					Name:         "test-ns",
					AutoTearDown: true,
				},
			},
		},
	}

	assert.Assert(t, phase.Run())
}

func TestCreateTestBase(t *testing.T) {
	r := &frame2.Run{
		T: t,
	}

	err := ConnectInitial()
	assert.Assert(t, err)

	testBase := NewTestBase("tb")

	phase := frame2.Phase{
		Runner: r,
		Setup: []frame2.Step{
			{
				Modify: &NamespaceCreateTestBase{
					Id:           "first",
					Kind:         "pub",
					TestBase:     testBase,
					AutoTearDown: true,
				},
			},
			{
				Modify: &NamespaceCreateTestBase{
					Id:           "second",
					Kind:         "pub",
					TestBase:     testBase,
					AutoTearDown: true,
				},
			},
		},
	}

	assert.Assert(t, phase.Run())

}
