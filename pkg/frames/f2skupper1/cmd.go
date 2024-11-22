package f2skupper1

import (
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

// If both Namespace and ClusterContext are empty, the command will be executed
// without --namespace
type CliSkupper struct {
	Args []string

	// The primary way to define the namespace
	Namespace string

	// Secondary way to get the namespace, used only if Namespace is empty
	F2Namespace *f2k8s.Namespace

	// You can configure any aspects of the command configuration.  However,
	// the fields Command, Args and Shell from the exec.Cmd element will be
	// cleared before execution.
	Cmd f2general.Cmd

	path string

	frame2.DefaultRunDealer
	frame2.Log
}

func (c CliSkupper) GetNamespace() string {
	if c.Namespace != "" {
		return c.Namespace
	}
	if c.F2Namespace != nil {
		return c.F2Namespace.GetNamespaceName()
	}
	return ""
}

func (c *CliSkupper) Validate() error {
	return c.Execute()
}

func (cs *CliSkupper) Execute() error {
	log.Printf("execute.CliSkupper %v", cs.Args)
	//	log.Printf("%#v", cs)
	baseArgs := []string{}

	// TODO change this when adding Podman to frame2
	baseArgs = append(baseArgs, "--platform", "kubernetes")

	if cs.F2Namespace != nil {
		file := cs.F2Namespace.GetKubeConfig().GetKubeconfigFile()
		if file != "" {
			baseArgs = append(baseArgs, "--kubeconfig", cs.F2Namespace.GetKubeConfig().GetKubeconfigFile())
		}
	}

	if cs.Namespace != "" {
		baseArgs = append(baseArgs, "--namespace", cs.Namespace)
	} else {
		if cs.F2Namespace != nil {
			baseArgs = append(baseArgs, "--namespace", cs.F2Namespace.GetNamespaceName())
		}
	}
	cmd := cs.Cmd
	cmd.Command = cs.path
	if cmd.Command == "" {
		cmd.Command = "skupper"
	}
	cmd.Cmd.Args = append(baseArgs, cs.Args...)

	err := cmd.Execute()
	if err != nil {
		log.Printf("CmdResult: %#v", cmd.CmdResult)
		return fmt.Errorf("execute.CliSkupper: %w", err)
	}
	return nil
}

func (c *CliSkupper) SetSkupperCliPath(path string, env []string) {
	c.path = path
	c.Cmd.AdditionalEnv = env
}
