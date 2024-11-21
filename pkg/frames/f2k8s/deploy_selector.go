package f2k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploySelector struct {
	Namespace Namespace
	Name      string
	Ctx       context.Context

	// Return value
	Deploy *appsv1.Deployment
}

func (d *DeploySelector) Execute() error {

	deploy, err := d.Namespace.DeploymentInterface().Get(d.Ctx, d.Name, metav1.GetOptions{})
	d.Deploy = deploy
	if err != nil {
		return err
	}
	return nil
}
