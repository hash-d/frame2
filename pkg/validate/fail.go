package validate

import (
	"errors"

	frame2 "github.com/hash-d/frame2/pkg"
)

type Fail struct {
	Reason string

	frame2.Log
	frame2.DefaultRunDealer
}

func (f Fail) Validate() error {
	return errors.New(f.Reason)
}
