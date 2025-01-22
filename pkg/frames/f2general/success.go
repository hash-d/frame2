package f2general

import frame2 "github.com/hash-d/frame2/pkg"

type Success struct {
	frame2.Log
	frame2.DefaultRunDealer
}

func (s Success) Execute() error {
	return nil
}

func (s Success) Validate() error {
	return nil
}
