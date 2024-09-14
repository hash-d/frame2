package k8sexecute

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretCreate struct {
	Namespace *f2k8s.Namespace
	Secret    *core.Secret

	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretCreate) Execute() error {

	_, err := s.Namespace.SecretInterface().Create(
		context.Background(),
		s.Secret,
		meta.CreateOptions{},
	)
	return err
}

type SecretDelete struct {
	Namespace *f2k8s.Namespace
	Name      string

	Secret *core.Secret // return

	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretDelete) Execute() error {

	var err error
	err = s.Namespace.SecretInterface().Delete(
		context.Background(),
		s.Name,
		meta.DeleteOptions{},
	)
	return err
}
