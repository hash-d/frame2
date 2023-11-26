package execute

import "github.com/skupperproject/skupper/test/utils/base"

type NginxDeploy struct {
	Namespace *base.ClusterContext
}

func (n NginxDeploy) Execute() error {
	return nil
}
