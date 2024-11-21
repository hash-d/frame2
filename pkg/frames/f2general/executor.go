package f2general

import frame2 "github.com/hash-d/frame2/pkg"

// This allows one to use an Executor as a validation, while ensuring that the
// test step explicitly shows it's using an Executor for validation.
//
// It's on the user to ensure whatever use of an executor they make will not
// have side effects, as validators are not expected to produce side effects.
//
// If an executor does produce a side effect, do not use this.  Put the
// executor on the Modify of the step (for the side effect, if desired), and
// use something else to validate.
//
// TODO: as an ergonomy option, add an "Executors []frame2.Executor" and an
// option for AND or OR on them?
type Executor struct {
	// Make sure your Executor has no side effects when using
	// validate.Executor
	Executor frame2.Executor

	frame2.Log
	frame2.DefaultRunDealer
}

func (e *Executor) Validate() error {
	p := frame2.Phase{
		Runner: e.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Modify: e.Executor,
			},
		},
	}
	return p.Run()
}
