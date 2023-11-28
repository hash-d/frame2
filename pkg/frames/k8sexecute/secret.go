package k8sexecute

import frame2 "github.com/hash-d/frame2/pkg"

type SecretCreate struct {
	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretCreate) Execute() error {
	return nil
}

type SecretDelete struct {
	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretDelete) Execute() error {
	return nil
}
