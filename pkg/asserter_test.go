package frame2_test

import (
	"errors"
	"fmt"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"gotest.tools/assert"
)

// Careful on the use of CheckError, as failing to set it when
// Errors are being checked may result in false positives/negatives.
//
// 50% of the bad uses are detected with a panic, the other 50%
// cannot be detected
type AsserterCheckErrorTesterItem struct {
	// An error to be checked with CheckError
	Error error

	// A boolean to be checked with Check()
	Condition   bool
	Message     string
	MessageArgs []any

	// By default, the boolean Condition is checked with Asserter.Check().
	// If CheckError is true, then the Error above is checked with
	// Asserter.CheckError()
	CheckError bool
}

type AsserterCheckErrorTester struct {
	Items     []AsserterCheckErrorTesterItem
	Failures  int
	Successes int

	ResultInError bool
}

func (a AsserterCheckErrorTester) Execute() error {
	asserter := frame2.Asserter{}

	for _, item := range a.Items {
		if item.CheckError {
			if item.Condition {
				panic("CheckError is true, but Condition was set to true.  Bad test setup")
			}
			asserter.CheckError(item.Error, item.Message, item.MessageArgs...)
		} else {
			if item.Error != nil {
				panic(fmt.Sprintf("Error (%v) is not nil, but no CheckError.  Bad test setup", item.Error))
			}
			asserter.Check(item.Condition, item.Message, item.MessageArgs...)
		}
	}

	if (asserter.Error() != nil) != a.ResultInError {
		return fmt.Errorf("Check failed: expected ResultInError %t, got %v", a.ResultInError, asserter.Error())
	}

	fail, success, total := asserter.GetStats()
	if fail+success != total {
		return fmt.Errorf(
			"%d failures + %d successes is different from the expected %d checks",
			fail, success, total,
		)
	}
	if fail != a.Failures {
		return fmt.Errorf(
			"expected %d failures, got %d",
			a.Failures, fail,
		)
	}
	// This will probably never be exercised, due to the two checks above,
	// but added for completeness
	if success != a.Successes {
		return fmt.Errorf(
			"expected %d successes, got %d",
			a.Successes, success,
		)
	}

	return nil
}

func TestAsserter(t *testing.T) {

	runner := frame2.Run{
		T: t,
	}
	phase := frame2.Phase{
		Runner: &runner,
		MainSteps: []frame2.Step{
			{
				Name: "Single Success",
				Modify: &AsserterCheckErrorTester{
					Items: []AsserterCheckErrorTesterItem{
						{
							Error:      nil,
							Message:    "This should not result in error",
							CheckError: true,
						},
					},
					ResultInError: false,
					Successes:     1,
				},
			},
			{
				Name: "Single failure",
				Modify: &AsserterCheckErrorTester{
					Items: []AsserterCheckErrorTesterItem{
						{
							Error:      fmt.Errorf("This is an error"),
							Message:    "This is an error",
							CheckError: true,
						},
					},
					ResultInError: true,
					Failures:      1,
				},
			},
			{
				Name: "Single true",
				Modify: &AsserterCheckErrorTester{
					Items: []AsserterCheckErrorTesterItem{
						{
							Condition: true,
							Message:   "This should not result in error",
						},
					},
					ResultInError: false,
					Successes:     1,
				},
			},
			{
				Name: "Single false",
				Modify: &AsserterCheckErrorTester{
					Items: []AsserterCheckErrorTesterItem{
						{
							Condition: false,
							Message:   "This is an error",
						},
					},
					ResultInError: true,
					Failures:      1,
				},
			},
			{
				Name: "Failed condition in the middle",
				Modify: &AsserterCheckErrorTester{
					Items: []AsserterCheckErrorTesterItem{
						{
							Error:      nil,
							Message:    "This is not an error",
							CheckError: true,
						},
						{
							Condition: false,
							Message:   "This is an error",
						},
						{
							Error:      nil,
							Message:    "This is not an error",
							CheckError: true,
						},
					},
					ResultInError: true,
					Failures:      1,
					Successes:     2,
				},
			},
			{
				Name: "Error in the middle",
				Modify: &AsserterCheckErrorTester{
					Items: []AsserterCheckErrorTesterItem{
						{
							Condition: true,
							Message:   "This is not an error",
						},
						{
							Error:      errors.New("Some error"),
							Message:    "This is an error",
							CheckError: true,
						},
						{
							Condition: true,
							Message:   "This is not an error",
						},
					},
					ResultInError: true,
					Failures:      1,
					Successes:     2,
				},
			},
		},
	}
	assert.Assert(t, phase.Run())

}
