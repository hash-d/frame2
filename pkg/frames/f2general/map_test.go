package f2general_test

import (
	"regexp"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"gotest.tools/assert"
)

func TestMapCheck(t *testing.T) {

	runner := frame2.Run{
		T: t,
	}

	testMap := map[string]string{
		"a": "A",
		"b": "B",
		"c": "C",
	}

	phase := frame2.Phase{
		Runner: &runner,
		MainSteps: []frame2.Step{
			{
				Name: "Positive",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						KeysPresent:             []string{"a", "b", "c"},
						KeysAbsent:              []string{"x", "y", "z"},
						Values:                  map[string]string{"a": "A", "c": "C"},
						ValuesOrMissing:         map[string]string{"b": "B", "z": "Z"},
						NegativeValues:          map[string]string{"a": "Z"},
						NegativeValuesOrMissing: map[string]string{"b": "Z", "z": "A"},
						RegexpValues:            map[string]regexp.Regexp{"a": *regexp.MustCompile("^A$")},
						MapType:                 "map",
					},
				},
			}, {
				Name: "Negative KeysPresent",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						KeysPresent: []string{"z"},
					},
				},
				ExpectError: true,
			}, {
				Name: "Negative KeysAbsent",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						KeysAbsent: []string{"a"},
					},
				},
				ExpectError: true,
			}, {
				Name: "Negative Values",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						Values: map[string]string{"a": "Z"},
					},
				},
				ExpectError: true,
			}, {
				Name: "Negative ValuesOrMissing",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						ValuesOrMissing: map[string]string{"a": "Z"},
					},
				},
				ExpectError: true,
			}, {
				Name: "Negative NegativeValues",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						NegativeValues: map[string]string{"a": "A"},
					},
				},
				ExpectError: true,
			}, {
				Name: "Negative NegativeValuesOrMissing",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						NegativeValuesOrMissing: map[string]string{"a": "A"},
					},
				},
				ExpectError: true,
			}, {
				Name: "Negative RegexpValues",
				Validator: MapCheckTester{
					Map: testMap,
					Checks: f2general.MapCheck{
						RegexpValues: map[string]regexp.Regexp{"a": *regexp.MustCompile("^asdf$")},
					},
				},
				ExpectError: true,
			},
		},
	}

	assert.Assert(t, phase.Run())

}

type MapCheckTester struct {
	Map    map[string]string
	Checks f2general.MapCheck

	frame2.DefaultRunDealer
	*frame2.Log
}

func (m MapCheckTester) Validate() error {
	return m.Checks.Check(m.Map)
}
