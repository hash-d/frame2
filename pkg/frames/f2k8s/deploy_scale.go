package f2k8s

import (
	"context"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeployScale struct {
	Namespace     *Namespace
	DeploymentGet DeploymentValidate // Do not populate the Namespace within the DeploymentGet; it will be auto-populated
	Replicas      int32
	Ctx           context.Context
}

func (d DeployScale) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)
	log.Printf("execute.DeployScale")

	d.DeploymentGet.Namespace = d.Namespace

	err := d.DeploymentGet.Validate()
	if err != nil {
		return err
	}

	deploy := d.DeploymentGet.Result

	deploy.Spec.Replicas = &d.Replicas
	_, err = d.Namespace.DeploymentInterface().Update(ctx, deploy, v1.UpdateOptions{})

	return err

}
