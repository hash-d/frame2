package frame2

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gotest.tools/assert"
)

var funcError = errors.New("FuncError")

type testChecks struct {
	input  []error
	result error
}

type test struct {
	config RetryOptions
	checks []testChecks
}

func TestRetry(t *testing.T) {
	table := []test{
		{
			// No retries, so the second item in the input should
			// never be returned
			config: RetryOptions{Interval: time.Millisecond},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Retries:  1,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   1,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Allow:    1,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Ignore:   2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError},
					result: funcError,
				}, {
					input:  []error{nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Allow:    2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, funcError},
					result: funcError,
				}, {
					input:  []error{nil, funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ensure:   2,
				Retries:  4,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, funcError, nil, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   2,
				Allow:    2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{nil, nil, funcError},
					result: funcError,
				}, {
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, nil, nil},
					result: nil,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, nil, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, nil, funcError, funcError, funcError},
					result: funcError,
				}, {
					input:  []error{nil, nil, funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Allow:    2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, funcError},
					result: funcError,
				},
			},
		}, {
			config: RetryOptions{
				Ignore:   2,
				Ensure:   2,
				Allow:    2,
				Retries:  2,
				Interval: time.Millisecond,
			},
			checks: []testChecks{
				{
					input:  []error{funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{nil, nil, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, nil, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, nil, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, nil, nil},
					result: nil,
				}, {
					input:  []error{funcError, funcError, funcError, funcError, funcError},
					result: funcError,
				},
			},
		},
	}

	for i, item := range table {
		t.Run(fmt.Sprintf("item-%d", i), func(t *testing.T) {
			for j, c := range item.checks {
				t.Run(fmt.Sprintf("check-%d", j), func(t *testing.T) {
					var n int
					_, err := Retry{
						Options: item.config,
						Fn: func() error {
							if n >= len(c.input) {
								t.Logf("Test failed: %+v", item)
								t.Fatalf("tried to access input #%d", n)
							}
							ret := c.input[n]
							n++
							return ret
						},
					}.Run()
					t.Logf("Got response %v", err)
					if !errors.Is(err, c.result) {
						t.Logf("Test failed: %+v", item)
						t.Errorf("%v != %v", err, c.result)
					}
					if n != len(c.input) {
						t.Logf("Test failed: %+v", item)
						t.Errorf("used %v items from total %v in input", n, len(c.input))
					}
				})
			}
		})

	}
}
func TestRetryOptMax(t *testing.T) {
	min := RetryOptions{}
	max := RetryOptions{
		Allow:      10,
		Ignore:     10,
		Ensure:     10,
		Retries:    10,
		Interval:   10,
		Quiet:      true,
		Min:        10,
		Rate:       10.0,
		KeepTrying: true,
		Ctx:        nil,
		Timeout:    10,
	}
	med := RetryOptions{
		Allow:      5,
		Ignore:     5,
		Ensure:     5,
		Retries:    5,
		Interval:   5,
		Quiet:      false,
		Min:        5,
		Rate:       5.0,
		KeepTrying: true,
		Ctx:        nil,
		Timeout:    5,
	}

	sumMinMax, cancel := min.Max(max)
	defer cancel()
	assert.Assert(t, sumMinMax == max)
	assert.Assert(t, &sumMinMax != &min, "the result should not point to one of its parameters")
	assert.Assert(t, &sumMinMax != &max, "the result should not point to one of its parameters")
	assert.Assert(t, min.Ctx == nil, "Max() should not affect its parameters")
	assert.Assert(t, max.Ctx == nil, "Max() should not affect its parameters")
	assert.Assert(t, sumMinMax.Ctx == nil, "If both parameters have nil contexts, a nil context should be returned")

	sumMaxMin, cancel := max.Max(min)
	defer cancel()
	assert.Assert(t, sumMaxMin == max)

	sumMinMed, cancel := min.Max(med)
	defer cancel()
	assert.Assert(t, sumMinMed == med)

	sumMedMin, cancel := med.Max(min)
	defer cancel()
	assert.Assert(t, sumMedMin == med)

	sumMedMax, cancel := med.Max(max)
	defer cancel()
	assert.Assert(t, sumMedMax == max)

	sumMaxMed, cancel := max.Max(med)
	defer cancel()
	assert.Assert(t, sumMaxMed == max)

	// Context checks
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	opt1 := RetryOptions{
		Ctx: ctx1,
	}
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	opt2 := RetryOptions{
		Ctx: ctx2,
	}
	twoCtx, cancelTwoCtx := opt1.Max(opt2)
	defer cancelTwoCtx()
	assert.Assert(t, twoCtx.Ctx.Err() == nil, "just confirming the merged context is not canceled")
	cancel1()

	// We have to wait for twoCtx to be closed in response of cancel1(), as that operation is
	// assynchronous, but we set a maximum wait time.  If that one is reached, the error will
	// be caught on the following assertion, not within the select block
	select {
	case <-twoCtx.Ctx.Done():
	case <-time.After(time.Second):
	}
	assert.Assert(t, twoCtx.Ctx.Err() != nil, "cancelling a component context should cancel the merged one")

	// We give ctx2 a second to ensure twoCtx' cancellation did not affect it
	select {
	case <-ctx2.Done():
	case <-time.After(time.Second):
	}
	assert.Assert(t, ctx2.Err() == nil, "cancelling the merged context should not cancel its components")

	// Check cancellation of first
	ctxA1, cancelA1 := context.WithCancel(context.Background())
	defer cancelA1()
	optA1 := RetryOptions{
		Ctx: ctxA1,
	}
	ctxA2, cancelA2 := context.WithCancel(context.Background())
	defer cancelA2()
	optA2 := RetryOptions{
		Ctx: ctxA2,
	}
	ctxA3, cancelA3 := context.WithCancel(context.Background())
	defer cancelA3()
	optA3 := RetryOptions{
		Ctx: ctxA3,
	}
	optA1A2, cancelA1A2 := optA1.Max(optA2)
	defer cancelA1A2()
	optA1A2A3, cancelA1A2A3 := optA1A2.Max(optA3)
	defer cancelA1A2A3()
	cancelA1()
	select {
	case <-optA1A2A3.Ctx.Done():
	case <-time.After(time.Second):
	}
	assert.Assert(t, optA1A2A3.Ctx.Err() != nil, "cancelling the 'first' context cancels the merged one")
	select {
	case <-ctxA2.Done():
	case <-ctxA3.Done():
	case <-time.After(time.Second):
	}
	assert.Assert(t, ctxA2.Err() == nil, "cancelling the merged context does not impact merged ones")
	assert.Assert(t, ctxA3.Err() == nil, "cancelling the merged context does not impact merged ones")
	assert.Assert(t, optA1A2.Ctx.Err() != nil, "cancelling the merged cancels the intermediary ones")

	// Check cancellation of last
	ctxB1, cancelB1 := context.WithCancel(context.Background())
	defer cancelB1()
	optB1 := RetryOptions{
		Ctx: ctxB1,
	}
	ctxB2, cancelB2 := context.WithCancel(context.Background())
	defer cancelB2()
	optB2 := RetryOptions{
		Ctx: ctxB2,
	}
	ctxB3, cancelB3 := context.WithCancel(context.Background())
	defer cancelB3()
	optB3 := RetryOptions{
		Ctx: ctxB3,
	}
	optB1B2, cancelB1B2 := optB1.Max(optB2)
	defer cancelB1B2()
	optB1B2B3, cancelB1B2B3 := optB1B2.Max(optB3)
	defer cancelB1B2B3()
	cancelB3()
	select {
	case <-optB1B2B3.Ctx.Done():
	case <-time.After(time.Second):
	}
	assert.Assert(t, optB1B2B3.Ctx.Err() != nil, "cancelling the 'last' context cancels the merged one")
	select {
	case <-ctxB1.Done():
	case <-ctxB2.Done():
	case <-time.After(time.Second):
	}
	assert.Assert(t, ctxB1.Err() == nil, "cancelling the merged context does not impact merged ones")
	assert.Assert(t, ctxB2.Err() == nil, "cancelling the merged context does not impact merged ones")
	assert.Assert(t, optB1B2.Ctx.Err() != nil, "cancelling the merged cancels the intermediary ones")

	// Two sources
	ctxC1, cancelC1 := context.WithCancel(context.Background())
	defer cancelC1()
	optC1 := RetryOptions{
		Ctx: ctxC1,
	}
	ctxC2, cancelC2 := context.WithCancel(context.Background())
	defer cancelC2()
	optC2 := RetryOptions{
		Ctx: ctxC2,
	}
	ctxC3, cancelC3 := context.WithCancel(context.Background())
	defer cancelC3()
	optC3 := RetryOptions{
		Ctx: ctxC3,
	}
	ctxC4, cancelC4 := context.WithCancel(context.Background())
	defer cancelC4()
	optC4 := RetryOptions{
		Ctx: ctxC4,
	}
	optC1C2, cancelC1C2 := optC1.Max(optC2)
	defer cancelC1C2()
	optC3C4, cancelC3C4 := optC3.Max(optC4)
	defer cancelC3C4()
	optC1C2C3C4, cancelC1C2C3C4 := optC1C2.Max(optC3C4)
	defer cancelC1C2C3C4()
	cancelC1()
	select {
	case <-optC1C2C3C4.Ctx.Done():
	case <-time.After(time.Second):
	}
	assert.Assert(t, optC1C2C3C4.Ctx.Err() != nil, "cancelling a member cancels the merged context")

}
