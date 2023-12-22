package execute_test

import (
	"errors"
	"log"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/validate"
)

func TestFunction(t *testing.T) {
	tests := frame2.Phase{
		Name: "TestFunction",
		MainSteps: []frame2.Step{
			{
				Name: "func-ok",
				Modify: execute.Function{
					Fn: func() error {
						log.Printf("Hello")
						return nil
					},
				},
			}, {
				Name: "func-fail",
				Validator: validate.Phase{
					// We need to use validate.Phase to capture the Modify error as a
					// validation step: Modify is supposed to always generate an actual
					// error, and ExpectError is only for validations.
					Phase: frame2.Phase{
						MainSteps: []frame2.Step{
							{
								Modify: execute.Function{
									Fn: func() error {
										return errors.New("failed!")
									},
								},
							},
						},
					},
				},
				ExpectError: true,
			},
		},
	}
	tests.RunT(t)
}
