package frame2

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const EnvFrame2Verbose = "SKUPPER_TEST_FRAME2_VERBOSE"

type Stepper interface {
	GetStep() *Step
}

type Step struct {
	Doc   string
	Name  string
	Level int
	// Whether the step should always print logs
	// Even if false, logs will be done if SKUPPER_TEST_FRAME2_VERBOSE
	Verbose      bool
	PreValidator Validator

	// TODO make this require a pointer, like validator does, to avoid
	// Runner disconnects
	Modify            Executor
	Validator         Validator
	Validators        []Validator
	ValidatorRetry    RetryOptions
	ValidatorFinal    bool // final validators are re-run at the test's end
	ValidatorSubFinal bool // and subfinal, at the end of subtests
	Substep           Stepper
	Substeps          []*Step
	SubstepRetry      RetryOptions
	// A simple way to invert the meaning of the Validator.  Validators
	// are encouraged to provide more specific negative testing behaviors,
	// but this serves for simpler testing.  If set, it inverts the
	// response from the call sent to Retry, so it can be used to wait
	// until an error is returned (but there is no control on which kind
	// of error that will be)
	// TODO: change this by OnError: Fail, Skip, Ignore, Expect?
	// TODO: add option to inspect error message for specific error; perhaps a function (error) bool,
	// function(error)error or function(error)frame2.action (where action is fail, skip, etc)
	ExpectError bool
	// TODO: ExpectIs, ExpectAs; use errors.Is, errors.As against a list of expected errors?
	SkipWhen bool
}

func (s *Step) GetStep() *Step {
	return s
}

func (s Step) Logf(format string, v ...interface{}) {
	if s.IsVerbose() {
		left := strings.Repeat(" ", s.Level)
		log.Printf(left+format, v...)
	}
}

func (s Step) IsVerbose() bool {
	return s.Verbose || os.Getenv(EnvFrame2Verbose) != ""
}

// Returns a list where Step.Validator is the first item, followed by
// Step.Validators
func (s Step) GetValidators() []Validator {
	validators := []Validator{}

	if s.Validator != nil {
		validators = append(validators, s.Validator)
	}

	validators = append(validators, s.Validators...)

	return validators
}

type TransformFunc func(any) (any, error)

// IterFrames will run transform() on each of its configured frames
// (Modify, Validator and Validators[]) and reassign the frame to
// the return of transform().
//
// It allows disruptors to inspect Executors and Validators in a
// single go.
func (s *Step) IterFrames(transform TransformFunc) error {

	if s.Modify != nil {
		if ret, err := transform(s.Modify); err == nil {
			if ret, ok := ret.(Executor); ok {
				s.Modify = ret
			} else {
				return fmt.Errorf("TransformFunc did not return an Executor on Modify (%T)", ret)
			}
		} else {
			return fmt.Errorf("TransformFunc returned error on Modify: %w", err)
		}
	}

	if s.Validator != nil {
		if ret, err := transform(s.Validator); err == nil {
			if ret, ok := ret.(Validator); ok {
				s.Validator = ret
			} else {
				return fmt.Errorf("TransformFunc did not return a Validator on Validator(%T)", ret)
			}
		} else {
			return fmt.Errorf("TransformFunc returned error on Validator: %w", err)
		}
	}

	for i, v := range s.Validators {
		if ret, err := transform(v); err == nil {
			if ret, ok := ret.(Validator); ok {
				s.Validators[i] = ret
			} else {
				return fmt.Errorf("TransformFunc did not return a Validator on Validators[%d](%T)", i, ret)
			}
		} else {
			return fmt.Errorf("TransformFunc returned error on Validators[%d]: %w", i, err)
		}
	}

	return nil
}

// type Validate struct {
// 	Validator
// 	// Every Validator runs inside a Retry.  If no options are given,
// 	// the default RetryOptions are used (ie, single run of Fn, with either
// 	// failed check or error causing the step to fail)
// 	RetryOptions
// }
//
// func (v Validate) GetRetryOptions() RetryOptions {
// 	return v.RetryOptions
// }

type Validator interface {
	Validate() error
	FrameLogger
}

// TODO: add some embedded items:
//  Namer adds record Name; allows individual validators to be named
//  Docer adds Doc; same as above
//  Skipper adds SkipWhen and IgnoreWhen, allowing individual validators to be
//  skipped programmatically

// TODO create ValidatorList, with Validator + RetryOptions

type Execute struct {
}

type Executor interface {
	Execute() error
}

type TearDowner interface {
	Teardown() Executor
}
