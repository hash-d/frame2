package f2k8s

import (
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
)

type NginxDeploy struct {
	Namespace     *Namespace
	Name          string            // default "nginx"
	Labels        map[string]string // default app: nginx
	SecretMount   []SecretMount
	ExposeService bool
	SkupperExpose bool // TODO
	Wait          time.Duration
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
				Modify: &DeploymentCreateSimple{
					Name:      name,
					Namespace: n.Namespace,
					Wait:      n.Wait,
					DeploymentOpts: DeploymentOpts{
						Image:        "ghcr.io/nginxinc/nginx-unprivileged:latest",
						Labels:       n.Labels,
						SecretMounts: n.SecretMount,
					},
				},
			}, {
				Modify: &ServiceCreate{
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
