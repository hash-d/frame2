package frame2

import (
	"fmt"
	"strings"
)

// A helper to check a list of conditions and report them back as a block
type Asserter struct {
	errors    []error
	failures  int
	successes int
	checks    int
}

// Checks the condition and returns a new error if false.
//
// The error is also saved on the list of errors in the Asserter, and the counters
// are updated
func (a *Asserter) Check(condition bool, message string, params ...any) error {
	a.checks += 1
	if condition {
		a.successes += 1
		return nil
	}
	a.failures += 1
	err := fmt.Errorf(message, params...)
	a.errors = append(a.errors, err)

	return err

}

// Updates the counters and error list, and return its input
func (a *Asserter) CheckError(err error, template string, msgParams ...any) error {
	a.checks += 1
	if err == nil {
		a.successes += 1
	} else {
		a.failures += 1
		msg := fmt.Sprintf(template, msgParams...)
		err = fmt.Errorf("%s: %w", msg, err)
		a.errors = append(a.errors, err)
	}
	return err
}

// Returns a new error that lists all errors detected by the Asserter, or nil
// if none.
func (a *Asserter) Error() error {
	if len(a.errors) > 0 {
		list := []string{}
		for _, i := range a.errors[1:] {
			list = append(list, fmt.Sprintf("%q", i.Error()))
		}
		failures := strings.Join(list, ", ")
		return fmt.Errorf(
			"Asserter failed: %d successes, %d failures.  First failure: %q; others: %s",
			a.successes,
			a.failures,
			a.errors[0].Error(),
			failures,
		)
	}
	return nil
}

func (a *Asserter) GetStats() (failures, successes, checks int) {
	return a.failures, a.successes, a.checks
}

func (a *Asserter) GetErrors() []error {
	return a.errors
}
