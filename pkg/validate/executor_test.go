package validate_test

import (
	"fmt"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/validate"
)

func TestExecutor(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	p := frame2.Phase{
		Runner: &r,
		MainSteps: []frame2.Step{
			{
				Name: "happy",
				Validator: &validate.Executor{
					Executor: execute.Function{
						Fn: func() error {
							return nil
						},
					},
				},
			}, {
				Name: "failures",
				Validators: []frame2.Validator{
					&validate.Executor{
						Executor: execute.Fail{},
					},
				},
				ExpectError: true,
			}, {
				Name: "more-failures",
				Validators: []frame2.Validator{
					&validate.Executor{
						Executor: execute.Function{
							Fn: func() error {
								return fmt.Errorf("This error was expected to happen")
							},
						},
					},
				},
				ExpectError: true,
			},
		},
	}
	p.Run()
}
