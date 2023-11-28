package k8svalidate

import (
	"bytes"
	"context"
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretGet struct {
	Namespace *base.ClusterContext
	Name      string

	// Look for expected contents of the secret (exact)
	Expect map[string][]byte

	// If true, the Expect map should be the full contents
	// of the secret.  If it is empty and ExpectAll is true,
	// for example, the secret must be empty.
	ExpectAll bool

	Secret *core.Secret

	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretGet) Validate() error {
	client := s.Namespace.VanClient.KubeClient

	var err error
	s.Secret, err = client.CoreV1().Secrets(s.Namespace.Namespace).Get(
		context.Background(),
		s.Name,
		meta.GetOptions{},
	)
	if err != nil {
		return err
	}
	if s.ExpectAll {
		if len(s.Expect) != len(s.Secret.Data) {
			return fmt.Errorf(
				"Secret %q has %d entries, different from expected %d",
				s.Name,
				len(s.Secret.Data),
				len(s.Expect),
			)
		}

	}
	for k, v := range s.Expect {
		actual, ok := s.Secret.Data[k]
		if !ok {
			return fmt.Errorf(
				"key %q not present on secret %q",
				k, s.Name,
			)
		}
		if !bytes.Equal(actual, v) {
			return fmt.Errorf(
				"key %q's value %q is different from expected %q on secret %q",
				k, actual, v, s.Name,
			)
		}
	}
	return err
}
