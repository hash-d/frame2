package f2general

import (
	"encoding/json"
	"fmt"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/jmespath/go-jmespath"
)

// TODO: perhaps ofer a number of Expressions that matches the JSON
// types?  Something like
//
// StringValue map[string]string
// IntValue map[string]int
// BoolValue map[string]bool
//
// Where the key is the search, and the value is the expected value.
//
// Need to work on that interface, but that would be a better match
type JSONMatcher struct {

	// The JMESPath expression to be executed.  There are
	// currently two way of using it.
	//
	// 1 - An expression that returns a list of booleans,
	//     such as
	//
	//     [?[0] == 'router'] |[].mode | map((&@ == 'edgee'), @)
	//
	//     In this case, all items are verified to be true,
	//     and the length checks are executed.
	//
	//     Use this to validate a value of the JSON structure
	//     has a certain value specified on the JMESPath
	//
	// 2 - An expression that returns a list of any other
	//     types, such as
	//
	//     [?[0] == 'sslProfile']
	//
	//     In this case, NotBoolList must be set to true,
	//     and only the length checks will be run.
	//
	Expression string

	// TODO
	// If set to True, the Expression is expected to return
	// a literal, and min/max/exact checks are not run.  If
	// false, the expression is expected to return a list
	Literal bool

	// If NotBoolList, the content checks are skipped, and
	// only the sizes are verified.
	NotBoolList bool

	// If Exact, Min and Max are all 0, we expect the
	// Expression to return a list with zero elements
	//
	// If Exact and Max are non-zero, only Max is checked
	Exact int

	// Min is inclusive (ie Min==1, then len() must be
	// at least 1)
	Min int

	// If you want any number of elements being returned
	// from the Expression, set Max to math.MaxInt
	//
	// Max is inclusive (ie, if Max=10, then len() must
	// be equal or less than 10)
	Max int

	Response interface{}
}

// Inspect a JSON structure using JMESPath
type JSON struct {
	// The JSON structure to be unmarshalled and analized, in
	// string format
	Data string

	// A list of verifications to be performed against the data
	Matchers []JSONMatcher

	*frame2.Log
}

func (j JSON) Validate() error {
	asserter := frame2.Asserter{}

	if j.Data == "" {
		return fmt.Errorf("f2general.JSON received empty data")
	}

	var data interface{}
	if err := json.Unmarshal([]byte(j.Data), &data); err != nil {
		return fmt.Errorf("unmarshaling of JSON data failed: %w", err)

	}

	for _, m := range j.Matchers {
		log.Printf("- Checking expression %q", m.Expression)
		var err error
		m.Response, err = jmespath.Search(m.Expression, data)
		if asserter.CheckError(err, "failed asserting JMESPath %q: %v", m.Expression, err) != nil {
			continue
		}
		log.Printf("  With result %v", m.Response)

		if !m.Literal {
			if l, ok := m.Response.([]interface{}); nil != asserter.Check(ok, "result of expression %q is not a list. Actual value: %+v (%T)", m.Expression, m.Response, m.Response) {
				continue
			} else {
				for i, item := range l {
					if item, ok := item.(bool); !m.NotBoolList && nil == asserter.Check(
						ok,
						"item #%d of expression %q is not a boolean (it's of type %T with value %v, instead)",
						i, m.Expression, item, item,
					) {
						asserter.Check(
							item,
							"item #%d of expression %q returned false",
							i, m.Expression,
						)
					}
				}
				length := len(l)
				if nil != asserter.Check(
					length >= m.Min,
					"expression %q did not match minimum number of elements %d (found %d)",
					m.Expression, m.Min, length,
				) {
					continue
				}
				if m.Max > 0 {
					if nil != asserter.Check(
						length <= m.Max,
						"expression %q returned %d elements, more than the configured maximum of %d",
						m.Expression, length, m.Max,
					) {
						continue
					}
				} else {
					if nil != asserter.Check(
						length == m.Exact,
						"expression %q returned %d elements, instead of the expected %d",
						m.Expression, length, m.Exact,
					) {
						continue
					}
				}

			}

		}
	}

	return asserter.Error()
}
