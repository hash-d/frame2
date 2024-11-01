package execute

import (
	"context"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeployScale struct {
	Namespace      f2k8s.Namespace
	DeploySelector // Do not populate the Namespace within the PodSelector; it will be auto-populated
	Replicas       int32
	Ctx            context.Context
}

func (d DeployScale) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)
	log.Printf("execute.DeployScale")

	d.DeploySelector.Namespace = d.Namespace

	err := d.DeploySelector.Execute()
	if err != nil {
		return err
	}

	deploy := d.DeploySelector.Deploy

	deploy.Spec.Replicas = &d.Replicas
	_, err = d.Namespace.DeploymentInterface().Update(ctx, deploy, v1.UpdateOptions{})

	return err

}
