package f2ocp

import (
	"context"
	"fmt"

	v1 "github.com/openshift/api/route/v1"
	clientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RouteGet struct {
	Namespace *base.ClusterContext

	Name string

	Ctx context.Context

	frame2.Log

	Result *v1.Route
}

func (r *RouteGet) Validate() error {
	ctx := frame2.ContextOrDefault(r.Ctx)

	var err error
	client, err := clientset.NewForConfig(r.Namespace.VanClient.RestConfig)
	if err != nil {
		return fmt.Errorf("failed to obtain clientset")
	}
	r.Result, err = client.Routes(r.Namespace.Namespace).Get(ctx, r.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	return nil
}
