package frame2

import (
	"testing"

	"gotest.tools/assert"
)

func TestInit(t *testing.T) {

	assert.Assert(t, GetId() != "")
	assert.Assert(t, GetShortId() != "")
}
