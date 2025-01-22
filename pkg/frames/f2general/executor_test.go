package f2general_test

import (
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
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
				Validator: &f2general.Executor{
					Executor: f2general.Function{
						Fn: func() error {
							return nil
						},
					},
				},
			}, {
				Name: "failures",
				Validators: []frame2.Validator{
					&f2general.Executor{
						Executor: f2general.Fail{},
					},
				},
				ExpectError: true,
			}, {
				Name: "more-failures",
				Validators: []frame2.Validator{
					&f2general.Executor{
						Executor: f2general.Function{
							Fn: func() error {
								return fmt.Errorf("this error was expected to happen")
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
