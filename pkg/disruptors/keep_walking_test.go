package disruptors_test

import (
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/disruptors"
	"gotest.tools/assert"
)

func TestKeepWalking(t *testing.T) {

	r := &frame2.Run{
		T: t,
		RequiredDisruptors: []frame2.Disruptor{
			&disruptors.KeepWalking{},
		},
	}
	defer r.Finalize()

	p0 := frame2.Phase{
		Runner: r,
		Doc:    "Execute validations that fail, and expect KeepWalking to intercept them and make the test ignore the failures",
		MainSteps: []frame2.Step{
			{
				Name: "Single Validator",
				Validator: &f2general.Executor{
					Executor: f2general.Fail{
						Reason: "This is failing, but Keep Walking should save it",
					},
				},
			}, {
				Name: "ExpectError should be unaffected",
				Validator: &f2general.Executor{
					Executor: f2general.Success{},
				},
				ExpectError: true,
			}, {
				Name: "On subtest",
				Substeps: []*frame2.Step{
					{
						Validator: &f2general.Executor{
							Executor: f2general.Fail{
								Reason: "Saved fail on unnamed substep",
							},
						},
					}, {
						Doc: "Following step should execute",
						Validator: &f2general.Executor{
							Executor: f2general.Success{},
						},
					},
				},
			}, {
				Doc: "Not on a named test",
				Validator: &f2general.Executor{
					Executor: f2general.Fail{
						Reason: "Saved fail on unnamed substep (and Final)",
					},
				},
				ValidatorFinal: true,
			}, {
				Name: "Last step: success",
				Validator: &f2general.Executor{
					Executor: f2general.Success{},
				},
			},
		},
	}

	assert.Assert(t, p0.Run())

}
