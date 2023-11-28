package k8sexecute

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretCreate struct {
	Namespace *base.ClusterContext
	Secret    *core.Secret

	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretCreate) Execute() error {
	client := s.Namespace.VanClient.KubeClient

	_, err := client.CoreV1().Secrets(s.Namespace.Namespace).Create(
		context.Background(),
		s.Secret,
		meta.CreateOptions{},
	)
	return err
}

type SecretDelete struct {
	Namespace *base.ClusterContext
	Name      string

	Secret *core.Secret // return

	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretDelete) Execute() error {
	client := s.Namespace.VanClient.KubeClient

	var err error
	err = client.CoreV1().Secrets(s.Namespace.Namespace).Delete(
		context.Background(),
		s.Name,
		meta.DeleteOptions{},
	)
	return err
}
