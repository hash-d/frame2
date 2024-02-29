package skupperexecute

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

type SkupperExpose struct {
	Namespace *base.ClusterContext
	Type      string
	Name      string

	// TODO.  Change this into some constants, so it can be reused and translated by different backing
	//        implementations.
	// A string that will compile into a Regex, which matches the command stderr to define an
	// expected failure.
	FailureReason string

	Address                string
	Headless               bool
	Protocol               string
	Ports                  []int
	PublishNotReadyAddress bool
	TargetPorts            []string
	EnableTls              bool // deprecated since 1.3
	GenerateTlsSecrets     bool

	AutoTeardown bool

	frame2.DefaultRunDealer
	execute.SkupperVersionerDefault
}

// Interface execute.SkupperUpgradable; allow this to be used with Upgrade disruptors
func (s SkupperExpose) SkupperUpgradable() *base.ClusterContext {
	return s.Namespace
}

func (s SkupperExpose) Execute() error {
	versions := []string{"1.2", "1.3"}
	target := s.WhichSkupperVersion(versions)
	var action frame2.Executor
	switch target {
	case "1.3", "":
		action = &SkupperExpose1_3{
			Namespace:              s.Namespace,
			Type:                   s.Type,
			Name:                   s.Name,
			FailureReason:          s.FailureReason,
			Address:                s.Address,
			Headless:               s.Headless,
			Protocol:               s.Protocol,
			Ports:                  s.Ports,
			PublishNotReadyAddress: s.PublishNotReadyAddress,
			TargetPorts:            s.TargetPorts,
			EnableTls:              s.EnableTls,
			GenerateTlsSecrets:     s.GenerateTlsSecrets,
		}
	case "1.2":
		action = &SkupperExpose1_2{
			Namespace:              s.Namespace,
			Type:                   s.Type,
			Name:                   s.Name,
			FailureReason:          s.FailureReason,
			Address:                s.Address,
			Headless:               s.Headless,
			Protocol:               s.Protocol,
			Ports:                  s.Ports,
			PublishNotReadyAddress: s.PublishNotReadyAddress,
			TargetPorts:            s.TargetPorts,
			EnableTls:              s.EnableTls || s.GenerateTlsSecrets, // 1.3 rename
		}
	default:
		panic("unnassigned version for CliSkupperInstall")
	}
	phase := frame2.Phase{
		Runner: s.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Modify: action,
			},
		},
	}

	return phase.Run()
}

// TODO: rename this to CLI; make a general type that can call
// the CLI, create annotations, use Ansible or site controller,
// per configuration.
type SkupperExpose1_3 struct {
	Namespace *base.ClusterContext
	Type      string
	Name      string

	// TODO.  Change this into some constants, so it can be reused and translated by different backing
	//        implementations.
	// A string that will compile into a Regex, which matches the command stderr to define an
	// expected failure.
	FailureReason string

	Address                string
	Headless               bool
	Protocol               string
	Ports                  []int
	PublishNotReadyAddress bool
	TargetPorts            []string
	EnableTls              bool
	GenerateTlsSecrets     bool

	AutoTeardown bool

	frame2.DefaultRunDealer
}

func (se SkupperExpose1_3) Execute() error {

	var args []string

	if se.Type == "" || se.Name == "" {
		return fmt.Errorf("SkupperExpose configuration error - type and name must be specified")
	}

	args = append(args, "expose", se.Type, se.Name)

	if se.Headless {
		args = append(args, "--headless")
	}

	if se.PublishNotReadyAddress {
		args = append(args, "--publish-not-ready-addresses")
	}

	if se.Address != "" {
		args = append(args, "--address", se.Address)
	}

	if se.Protocol != "" {
		args = append(args, "--protocol", se.Protocol)
	}

	if len(se.TargetPorts) != 0 {
		args = append(args, "--target-port", strings.Join(se.TargetPorts, ","))
	}

	if len(se.Ports) != 0 {
		var tmpPorts []string
		for _, p := range se.Ports {
			tmpPorts = append(tmpPorts, strconv.Itoa(p))
		}
		args = append(args, "--port", strings.Join(tmpPorts, ","))
	}

	if se.EnableTls {
		args = append(args, "--enable-tls")
	}
	if se.GenerateTlsSecrets {
		args = append(args, "--generate-tls-secrets")
	}

	cmd := execute.Cmd{}

	if se.FailureReason != "" {
		cmd.FailReturn = []int{0}
		re, err := regexp.Compile(se.FailureReason)
		if err != nil {
			return fmt.Errorf("SkupperExpose failed to compile FailureReason %q as a regexp: %w", se.FailureReason, err)
		}
		cmd.Expect = cli.Expect{
			StdErrRe: []regexp.Regexp{*re},
		}
	}

	phase := frame2.Phase{
		Runner: se.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: se.Namespace,
					Cmd:            cmd,
				},
			},
		},
	}

	return phase.Run()
}

func (se SkupperExpose1_3) Teardown() frame2.Executor {

	if !se.AutoTeardown {
		return nil
	}

	return SkupperUnexpose{
		Namespace: se.Namespace,
		Type:      se.Type,
		Name:      se.Name,
		Address:   se.Address,

		Runner: se.Runner,
	}

}

// TODO: rename this to CLI; make a general type that can call
// the CLI, create annotations, use Ansible or site controller,
// per configuration.
type SkupperExpose1_2 struct {
	Namespace *base.ClusterContext
	Type      string
	Name      string

	// TODO.  Change this into some constants, so it can be reused and translated by different backing
	//        implementations.
	// A string that will compile into a Regex, which matches the command stderr to define an
	// expected failure.
	FailureReason string

	Address                string
	Headless               bool
	Protocol               string
	Ports                  []int
	PublishNotReadyAddress bool
	TargetPorts            []string
	EnableTls              bool

	AutoTeardown bool

	frame2.DefaultRunDealer
}

func (se SkupperExpose1_2) Execute() error {

	var args []string

	if se.Type == "" || se.Name == "" {
		return fmt.Errorf("SkupperExpose configuration error - type and name must be specified")
	}

	args = append(args, "expose", se.Type, se.Name)

	if se.Headless {
		args = append(args, "--headless")
	}

	if se.PublishNotReadyAddress {
		args = append(args, "--publish-not-ready-addresses")
	}

	if se.Address != "" {
		args = append(args, "--address", se.Address)
	}

	if se.Protocol != "" {
		args = append(args, "--protocol", se.Protocol)
	}

	if len(se.TargetPorts) != 0 {
		args = append(args, "--target-port", strings.Join(se.TargetPorts, ","))
	}

	if len(se.Ports) != 0 {
		var tmpPorts []string
		for _, p := range se.Ports {
			tmpPorts = append(tmpPorts, strconv.Itoa(p))
		}
		args = append(args, "--port", strings.Join(tmpPorts, ","))
	}

	if se.EnableTls {
		args = append(args, "--enable-tls")
	}

	cmd := execute.Cmd{}

	if se.FailureReason != "" {
		cmd.FailReturn = []int{0}
		re, err := regexp.Compile(se.FailureReason)
		if err != nil {
			return fmt.Errorf("SkupperExpose failed to compile FailureReason %q as a regexp: %w", se.FailureReason, err)
		}
		cmd.Expect = cli.Expect{
			StdErrRe: []regexp.Regexp{*re},
		}
	}

	phase := frame2.Phase{
		Runner: se.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: se.Namespace,
					Cmd:            cmd,
				},
			},
		},
	}

	return phase.Run()
}

func (se SkupperExpose1_2) Teardown() frame2.Executor {

	if !se.AutoTeardown {
		return nil
	}

	return SkupperUnexpose{
		Namespace: se.Namespace,
		Type:      se.Type,
		Name:      se.Name,
		Address:   se.Address,

		Runner: se.Runner,
	}

}

type SkupperUnexpose struct {
	Namespace *base.ClusterContext
	Type      string
	Name      string
	Address   string

	Runner *frame2.Run
}

func (su SkupperUnexpose) Execute() error {
	var args []string

	if su.Type == "" || su.Name == "" {
		return fmt.Errorf("SkupperExpose configuration error - type and name must be specified")
	}

	args = append(args, "unexpose", su.Type, su.Name)

	if su.Address != "" {
		args = append(args, "--address", su.Address)
	}

	phase := frame2.Phase{
		Runner: su.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: su.Namespace,
				},
			},
		},
	}

	return phase.Run()

}
