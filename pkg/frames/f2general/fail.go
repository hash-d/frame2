package f2general

import (
	"fmt"
	frame2 "github.com/hash-d/frame2/pkg"
)

type Fail struct {
	Reason string

	frame2.Log
	frame2.DefaultRunDealer
}

func (f Fail) Execute() error {
	return fmt.Errorf("failed as requested (%q)", f.Reason)
}

func (f Fail) Validate() error {
	return fmt.Errorf("failed as requested (%q)", f.Reason)
}
