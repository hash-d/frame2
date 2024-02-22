package frame2

import (
	"fmt"

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
type CauseEffect[T Executor] struct {
	Patch      T
	Validators []Validator
}

// A generic cause-effect verifier
//
// On each cycle, the Setup steps will first be executed, then
// the BaseFrame will be patched with the Patch of the current
// CauseEffect and executed.  Next, its Validators will be run
//
// The Setup is per-cycle.  If additional, multi-cycle setup is
// required, it's up to the enclosing test to provide it.
type Effects[T Executor] struct {
	Name      string
	BaseFrame T
	Setup     []Step
	Effects   map[string]CauseEffect[T]
	TearDown  []Step

	// Each combo contains a list of Effects.  After all
	// of their patches are applied to it, the BaseFrame
	// is run and all of their validators are run.
	Combos map[string][]string

	ExecutionProfile ExecutionProfile
}

// Stepper
func (e *Effects[T]) GetStep() *Step {
	var s *Step
	switch e.ExecutionProfile {
	case COMBO:
		s = e.getComboStep()
	case INDIVIDUAL:
		s = e.getIndividualStep()
	case BOTH:
		s = e.getBothStep()
	default:
		panic(fmt.Sprintf("no such ExecutionProfile: %v", e.ExecutionProfile))
	}
	return s
}

func (e Effects[T]) getComboStep() *Step {
	s := Step{
		Name: e.Name,
	}
	for name, effects := range e.Combos {
		frame := e.BaseFrame
		validators := []Validator{}
		for _, effect := range effects {
			error := mergo.Merge(&frame, e.Effects[effect].Patch)
			if error != nil {
				panic("error merging structs")
			}
			validators = append(validators, e.Effects[effect].Validators...)
		}
		sub := Step{
			Name:       name,
			Modify:     frame,
			Validators: validators,
		}
		s.Substeps = append(s.Substeps, &sub)
	}
	return &s
}

func (e Effects[T]) getIndividualStep() *Step {
	s := Step{
		Name: e.Name,
	}

	for name, effect := range e.Effects {
		frame := e.BaseFrame
		error := mergo.Merge(&frame, effect.Patch)
		if error != nil {
			panic("error merging structs")
		}
		sub := Step{
			Name: name,
			Modify: Phase{
				MainSteps: []Step{
					{
						Modify:     frame,
						Validators: effect.Validators,
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
func (e Effects[T]) getBothStep() *Step {
	s := e.getComboStep()
	i := e.getIndividualStep()
	s.Substeps = append(s.Substeps, i.Substeps...)
	return s
}
