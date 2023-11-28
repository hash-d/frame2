package execute

import (
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/k8s"
)

type NginxDeploy struct {
	Namespace     *base.ClusterContext
	Name          string            // default "nginx"
	Labels        map[string]string // default app: nginx
	SecretMount   []k8s.SecretMount
	ExposeService bool
	SkupperExpose bool // TODO
}

func (n NginxDeploy) Execute() error {
	name := n.Name
	if name == "" {
		name = "nginx"
	}
	if len(n.Labels) == 0 {
		n.Labels = map[string]string{"app": "nginx"}
	}
	p := frame2.Phase{
		MainSteps: []frame2.Step{
			{
				Modify: &K8SDeploymentOpts{
					Name:      name,
					Namespace: n.Namespace,
					Wait:      2 * time.Minute,
					DeploymentOpts: k8s.DeploymentOpts{
						Image:        "ghcr.io/nginxinc/nginx-unprivileged:latest",
						Labels:       n.Labels,
						SecretMounts: n.SecretMount,
					},
				},
			}, {
				Modify: &K8SServiceCreate{
					Namespace: n.Namespace,
					Name:      name,
					Selector:  n.Labels,
					Ports:     []int32{8080},
				},
				SkipWhen: !n.ExposeService,
			},
		},
	}

	return p.Run()
}
