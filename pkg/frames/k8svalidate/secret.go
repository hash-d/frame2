package k8svalidate

import (
	"bytes"
	"context"
	"fmt"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Uniformize fields and struct name, between this and ConfigMap
type SecretGet struct {
	Namespace *f2k8s.Namespace
	Name      string

	// TODO change all these for a f2general.MapCheck
	// Look for expected contents of the secret (exact)
	Expect map[string][]byte

	// If true, the Expect map should be the full contents
	// of the secret.  If it is empty and ExpectAll is true,
	// for example, the secret must be empty.
	ExpectAll bool

	// Checks that all listed keys are present on the
	// secret, regardless of their values
	Keys []string

	// Checks that listed keys are _not_ present on the
	// Secret
	AbsentKeys []string

	// If set, the Secret is expected to not be present;
	// fail if it exists
	ExpectAbsent bool

	Labels      f2general.MapCheck
	Annotations f2general.MapCheck

	// Function to run against the Secret, to validate.  Provides a way
	// to execute more complex validations not available above, inline
	SecretValidator func(corev1.Secret) error

	Secret *core.Secret

	*frame2.Log
	frame2.DefaultRunDealer
}

func (s SecretGet) Validate() error {
	var err error
	s.Secret, err = s.Namespace.SecretInterface().Get(
		context.Background(),
		s.Name,
		meta.GetOptions{},
	)
	if err != nil {
		if s.ExpectAbsent {
			return nil
		}
		return err
	}
	if s.ExpectAbsent {
		return fmt.Errorf("secret %q was expected be absent, but was found in namespace %q", s.Name, s.Namespace.GetNamespaceName())
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
		log.Printf("- checking key %q for its value", k)
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
	for _, k := range s.Keys {
		log.Printf("- checking key %q for its presence", k)
		if _, ok := s.Secret.Data[k]; !ok {
			return fmt.Errorf(
				"key %q not present on secret %q",
				k, s.Name,
			)
		}
	}
	for _, k := range s.AbsentKeys {
		log.Printf("- checking key %q for its absence", k)
		if _, ok := s.Secret.Data[k]; ok {
			return fmt.Errorf(
				"key %q should not be present on secret %q",
				k, s.Name,
			)
		}
	}
	asserter := frame2.Asserter{}
	if s.Labels.MapType == "" {
		s.Labels.MapType = "label"
	}
	asserter.CheckError(s.Labels.Check(s.Secret.Labels), "label verification")
	if s.Annotations.MapType == "" {
		s.Annotations.MapType = "annotation"
	}
	asserter.CheckError(s.Annotations.Check(s.Secret.Annotations), "annotation verification")

	if s.SecretValidator != nil {
		log.Printf("- Running SecretValidator")
		asserter.CheckError(s.SecretValidator(*s.Secret), "SecretValidator failed")
	}
	return asserter.Error()
}
