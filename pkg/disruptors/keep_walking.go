package disruptors

import (
	frame2 "github.com/hash-d/frame2/pkg"
)

// Ignore any validator failures, and just keep going
// no mater what.
//
// TODO:
//   - Make it configurable: Validator, Modifier, both
//   - Make it configurable: select items per ID
//     (such as SubT0.m0.p0.s1.m0.p0.s0.m0.p0.s0.m0)
type KeepWalking struct {
}

func (k KeepWalking) DisruptorEnvValue() string {
	return "KEEP_WALKING"
}

func (k KeepWalking) ValidationResultHook(runner *frame2.Run, step frame2.Step, err error) error {

	if err != nil {
		step.Logf("KEEP_WALKING: ignore failure on step %v", step.Doc)
	}

	return nil
}
