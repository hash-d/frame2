package f2general_test

import (
	"errors"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"log"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
)

func TestFunction(t *testing.T) {
	runner := &frame2.Run{
		T: t,
	}

	tests := frame2.Phase{
		Runner: runner,
		Name:   "TestFunction",
		MainSteps: []frame2.Step{
			{
				Name: "func-ok",
				Modify: f2general.Function{
					Fn: func() error {
						log.Printf("Hello")
						return nil
					},
				},
			}, {
				Name: "func-fail",
				Validator: f2general.Phase{
					// We need to use validate.Phase to capture the Modify error as a
					// validation step: Modify is supposed to always generate an actual
					// error, and ExpectError is only for validations.
					Phase: frame2.Phase{
						Doc: "Dummy phase",
						MainSteps: []frame2.Step{
							{
								Modify: f2general.Function{
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
	tests.Run()
}
