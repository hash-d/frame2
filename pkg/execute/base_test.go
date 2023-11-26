package execute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
)

func TestBuildClusterContext(t *testing.T) {
	r := frame2.Run{
		T: t,
	}
	p := frame2.Phase{
		Runner: &r,
	}
	p.Run()
}
