package subrunner

import (
	"context"
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/imdario/mergo"
)

type ExecutionProfile int

const (
	COMBO ExecutionProfile = iota
	INDIVIDUAL
	BOTH
	// RANDOM: random apply
)

// On each cycle executed by Effects, its BaseFrame is patched
// with the Patch below, and the Validators ensue
type CauseEffect[T frame2.Executor] struct {
	Doc             string
	Patch           T
	Validators      []frame2.Validator
	ValidatorsRetry frame2.RetryOptions
}

// A generic cause-effect verifier
//
// On each cycle, the Setup steps will first be executed, then
// the BaseFrame will be patched with the Patch of the current
// CauseEffect and executed.  Next, its Validators will be run
//
// The Setup is per-cycle.  If additional, multi-cycle setup is
// required, it's up to the enclosing test to provide it.
type Effects[T frame2.Executor, PT interface {
	*T
	Execute() error
}] struct {
	Name      string
	Doc       string
	BaseFrame PT
	Setup     []frame2.Step
	Effects   map[string]CauseEffect[T]
	TearDown  []frame2.Step

	// Each combo contains a list of Effects.  After all
	// of their patches are applied to it, the BaseFrame
	// is run and all of their validators are run.
	Combos map[string][]string

	ExecutionProfile ExecutionProfile
}

// Stepper
func (e *Effects[T, PT]) GetStep() (*frame2.Step, context.CancelFunc) {
	var s *frame2.Step
	var cancel context.CancelFunc
	switch e.ExecutionProfile {
	case COMBO:
		s, cancel = e.getComboStep()
	case INDIVIDUAL:
		s = e.getIndividualStep()
	case BOTH:
		s, cancel = e.getBothStep()
	default:
		panic(fmt.Sprintf("no such ExecutionProfile: %v", e.ExecutionProfile))
	}
	return s, cancel
}

func (e Effects[T, PT]) getComboStep() (*frame2.Step, context.CancelFunc) {
	var cancel context.CancelFunc
	s := frame2.Step{
		Name: e.Name,
		Doc:  e.Doc,
	}
	for name, effects := range e.Combos {
		frame := *e.BaseFrame
		validators := []frame2.Validator{}
		opt := frame2.RetryOptions{}
		for _, effect := range effects {
			error := mergo.Merge(&frame, e.Effects[effect].Patch)
			if error != nil {
				panic("error merging structs")
			}
			validators = append(validators, e.Effects[effect].Validators...)
			opt, cancel = opt.Max(e.Effects[effect].ValidatorsRetry)
		}
		sub := frame2.Step{
			Name: name,
			Modify: frame2.Phase{
				MainSteps: []frame2.Step{
					{
						Modify:            PT(&frame),
						Validators:        validators,
						ValidatorSubFinal: true,
						ValidatorRetry:    opt,
					},
				},
				Teardown: e.TearDown,
			},
		}
		s.Substeps = append(s.Substeps, &sub)
	}
	return &s, cancel
}

func (e Effects[T, PT]) getIndividualStep() *frame2.Step {
	s := frame2.Step{
		Name: e.Name,
		Doc:  e.Doc,
	}

	for name, effect := range e.Effects {
		frame := *e.BaseFrame
		err := mergo.Merge(&frame, effect.Patch)
		if err != nil {
			panic(fmt.Sprintf("error merging structs (%s)", err))
		}
		sub := frame2.Step{
			Name: name,
			Modify: frame2.Phase{
				MainSteps: []frame2.Step{
					{
						Modify:            PT(&frame),
						Validators:        effect.Validators,
						ValidatorSubFinal: true,
						ValidatorRetry:    effect.ValidatorsRetry,
					},
				},
				Teardown: e.TearDown,
			},
		}
		s.Substeps = append(s.Substeps, &sub)
	}

	return &s
}

// Returns a step with substeps for both Combo and Individual
// tests (in that order)
func (e Effects[T, PT]) getBothStep() (*frame2.Step, context.CancelFunc) {
	s, cancel := e.getComboStep()
	i := e.getIndividualStep()
	s.Substeps = append(s.Substeps, i.Substeps...)
	return s, cancel
}

func (e Effects[T, PT]) GetPhase(runner *frame2.Run) frame2.Phase {
	steps, cancel := e.GetStep()
	phase := frame2.Phase{
		Runner: runner,
		MainSteps: []frame2.Step{
			*steps,
		},
	}
	// We cannot defer cancel(), as we're just returning a phase, not
	// running it.  The context cancellation needs to be done after
	// the steps are actually executed, so we add that as a step to
	// the returned phase
	if cancel != nil {
		phase.MainSteps = append(phase.MainSteps, frame2.Step{
			Doc: "cancel merged context",
			Modify: &execute.Function{
				Fn: func() error {
					cancel()
					return nil
				},
			},
		})
	}
	return phase
}
