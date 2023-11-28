package k8svalidate

import frame2 "github.com/hash-d/frame2/pkg"

type SecretGet struct {
	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretGet) Validate() error {
	return nil
}
