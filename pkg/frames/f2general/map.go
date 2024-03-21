package f2general

import (
	"fmt"
	"log"
	"regexp"

	frame2 "github.com/hash-d/frame2/pkg"
)

type MapCheck struct {

	// List of keys expected to be present on the map, regardless of value
	KeysPresent []string

	// List of keys expected to _not_ be present on the map
	KeysAbsent []string

	// Keys that are expected to be present, with specific values
	Values map[string]string

	// If key is present, it must have the specific value.  A missing key is
	// considered ok
	ValuesOrMissing map[string]string

	// Annotations listed on this map must not have the mapped value.  If the
	// key does not exist at all on the map, the NegativeAnnotations test is
	// considered successful, unless NegativeAnnotationsExist is set to true
	NegativeValues map[string]string

	// If key is present, it must not have the specific value.  A missing key is
	// considered ok
	NegativeValuesOrMissing map[string]string

	// Match regexp against values
	RegexpValues map[string]regexp.Regexp

	MapValidator func(map[string]string) error

	// This is used informationally, on error messages.  It can be anything, such
	// as 'label' or 'annotation'
	MapType string
}

func (mc MapCheck) Check(m map[string]string) error {
	asserter := frame2.Asserter{}

	kind := fmt.Sprintf("%s ", mc.MapType)
	if kind == " " {
		kind = ""
	}

	for _, k := range mc.KeysPresent {
		_, ok := m[k]
		log.Printf("- checking for presence of key %q", k)
		asserter.Check(ok, "%skey %q not found", mc.MapType, k)
	}
	for _, k := range mc.KeysAbsent {
		log.Printf("- checking for absence of key %q", k)
		_, ok := m[k]
		asserter.Check(!ok, "%skey %q found, unexpectedly", mc.MapType, k)
	}
	for k, v := range mc.Values {
		log.Printf("- checking for key %q=%q", k, v)
		mapValue, ok := m[k]
		if asserter.Check(ok, "%skey %q not found", mc.MapType, k) == nil {
			asserter.Check(
				v == mapValue,
				"%skey %q has value %q, while expected was %q",
				mc.MapType, k, mapValue, v,
			)
		}
	}
	for k, v := range mc.ValuesOrMissing {
		log.Printf("- checking for key %q=%q, or missing", k, v)
		if mapValue, ok := m[k]; ok {
			asserter.Check(
				v == mapValue,
				"%skey %q has value %q, while expected was %q",
				mc.MapType, k, mapValue, v,
			)
		}
	}
	for k, v := range mc.NegativeValues {
		log.Printf("- checking for key %q!=%q", k, v)
		mapValue, ok := m[k]
		if asserter.Check(ok, "%skey %q not found", mc.MapType, k) == nil {
			asserter.Check(
				v != mapValue,
				"%skey %q has unexpected value %q",
				mc.MapType, k, mapValue,
			)
		}
	}
	for k, v := range mc.NegativeValuesOrMissing {
		log.Printf("- checking for key %q!=%q, or missing", k, v)
		if mapValue, ok := m[k]; ok {
			asserter.Check(
				v != mapValue,
				"k%sey %q has unexpected value %q",
				mc.MapType, k, mapValue,
			)
		}
	}
	for k, v := range mc.RegexpValues {
		log.Printf("- checking for key %q matching regex %v", k, v)
		mapValue, ok := m[k]
		if asserter.Check(ok, "key %q not found", k) == nil {
			asserter.Check(
				v.MatchString(mapValue),
				"%skey %q has unexpected value %q (did not match regexp %v)",
				mc.MapType, k, mapValue, v,
			)
		}
	}

	if mc.MapValidator != nil {
		log.Printf("- Running MapValidator")
		asserter.CheckError(mc.MapValidator(m), "MapValidator failed")
	}

	return asserter.Error()
}

func (mc MapCheck) CheckBytes(m map[string][]byte) error {
	strMap := map[string]string{}
	for k, v := range m {
		strMap[k] = string(v)
	}
	return mc.Check(strMap)
}
