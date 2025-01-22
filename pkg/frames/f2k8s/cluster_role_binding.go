package f2k8s

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterRoleBindingGet struct {
	// Even if not a namespaced resource, the Namespace is
	// required, so the frame can have access to the KubeClient
	//
	// The Namespace will then also be used to confirm that the
	// acquired binding points to this namespace
	Namespace *Namespace
	Name      string
	Ctx       context.Context

	ClusterRole string

	ExpectAbsent bool

	Result *rbacv1.ClusterRoleBinding

	frame2.Log
	frame2.DefaultRunDealer
}

func (c *ClusterRoleBindingGet) Validate() error {
	ctx := frame2.ContextOrDefault(c.Ctx)
	asserter := frame2.Asserter{}
	bindings, err := c.Namespace.ClusterRoleBindingInterface().Get(ctx, c.Name, v1.GetOptions{})
	c.Result = bindings
	if err != nil {
		if c.ExpectAbsent && errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	var foundSubject bool
	for _, subject := range bindings.Subjects {
		if c.Namespace.GetNamespaceName() == subject.Namespace {
			foundSubject = true
		}
	}
	asserter.Check(foundSubject, "no subject with namespace %q for clusterrolebinding %q", bindings.Namespace, c.Name)
	if c.ClusterRole != "" {
		asserter.Check(bindings.RoleRef.Name == c.ClusterRole, "clusterrolebinding points to unexpected role %q", bindings.RoleRef.Name)
	}
	return asserter.Error()
}
