package frame2

import (
	"os"
	"testing"

	"golang.org/x/exp/slices"
	"gotest.tools/assert"
)

// SourceRoot is sensible to refactoring, so we make sure
// it still points to the actual root if moved.
func TestSourceRoot(t *testing.T) {

	p := SourceRoot()
	files, err := os.ReadDir(p)
	assert.Assert(t, err)

	present := []string{}
	for _, f := range files {
		present = append(present, f.Name())
	}

	asserter := Asserter{}

	expected_components := []string{"go.mod", "go.sum", "pkg"}
	for _, c := range expected_components {
		asserter.Check(slices.Contains(present, c), "root directory is expected to contain child %q", c)
	}
	assert.Assert(t, asserter.Error())

}
