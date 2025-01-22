package f2general

import frame2 "github.com/hash-d/frame2/pkg"

// This frame exists to test frame2 itself
//
// You should be able to use it to do complex validations,
// but you're probably better off using a simple list of
// Validators.
//
// On the provided frame2.Phase, do not set the Runner:
// it will be overridden
type Phase struct {
	Phase frame2.Phase
	*frame2.Log
	frame2.DefaultRunDealer
}

func (p Phase) Validate() error {
	p.Phase.Runner = p.Runner
	return p.Phase.Run()

}
