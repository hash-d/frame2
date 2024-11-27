package f2skupper1

import (
	"context"
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1const"
	"strings"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

// A Skupper installation that uses some default configurations.
// It cannot be configured.  For a configurable version, use
// SkupperInstall, instead.
type SkupperInstallSimple struct {
	Namespace     *f2k8s.Namespace
	EnableConsole bool

	frame2.DefaultRunDealer
}

func (sis SkupperInstallSimple) Execute() error {
	phase := frame2.Phase{
		Runner: sis.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupperInstall{
					Namespace:           sis.Namespace,
					EnableConsole:       sis.EnableConsole,
					EnableFlowCollector: sis.EnableConsole,
				},
			},
		},
	}
	return phase.Run()
}

type CliSkupperInstall struct {
	Namespace                *f2k8s.Namespace
	Ctx                      context.Context
	MaxWait                  time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait                 bool
	SkipStatus               bool
	EnableConsole            bool
	EnableFlowCollector      bool
	Annotations              []string
	CreateNetworkPolicy      bool
	EnableClusterPermissions bool
	SiteName                 string
	RouterLogging            string
	RouterMode               string
	Ingress                  string
	IngressHost              string
	DisableServiceSync       bool
	RouterCPU                string

	ConsoleAuth     string
	ConsoleUser     string
	ConsolePassword string

	frame2.DefaultRunDealer
	SkupperVersionerDefault
}

// TODO: replace this by f2k8s.Namespace
func (c CliSkupperInstall) GetNamespace() string {
	return c.Namespace.GetNamespaceName()
}

// Interface execute.SkupperUpgradable; allow this to be used with Upgrade disruptors
func (c CliSkupperInstall) SkupperUpgradable() *f2k8s.Namespace {
	return c.Namespace
}

func (s CliSkupperInstall) Execute() error {

	versions := []string{"1.2", "1.3"}
	target := s.WhichSkupperVersion(versions)

	args := []string{"init"}

	// EnableConsole
	switch target {
	case "1.3", "":
		if s.EnableConsole {
			args = append(args, "--enable-console")
		}
	case "1.2":
		// On 1.3 the default changed from --enable-console=true to --enable-console=false.
		// For this reason, on 1.2 we need to always specify the console flag.
		args = append(args, fmt.Sprintf("--enable-console=%t", s.EnableConsole))
	}

	// EnableFlowColector
	switch target {
	case "1.3", "":
		if s.EnableFlowCollector {
			args = append(args, "--enable-flow-collector")
		}
	case "1.2":
		// TODO: make this configurable, so it can be just ignored instead of
		//       failing every time.
		return fmt.Errorf("flow collector not available for version <1.3")
	}

	if s.DisableServiceSync {
		args = append(args, "--enable-service-sync=false")
	}

	if s.ConsoleAuth != "" {
		args = append(args, fmt.Sprintf("--console-auth=%s", s.ConsoleAuth))
	}
	if s.ConsoleUser != "" {
		args = append(args, fmt.Sprintf("--console-user=%s", s.ConsoleUser))
	}
	if s.ConsolePassword != "" {
		args = append(args, fmt.Sprintf("--console-password=%s", s.ConsolePassword))
	}
	if s.SiteName != "" {
		args = append(args, fmt.Sprintf("--site-name=%s", s.SiteName))
	}
	if s.RouterLogging != "" {
		args = append(args, fmt.Sprintf("--router-logging=%s", s.RouterLogging))
	}
	if s.RouterMode != "" {
		args = append(args, fmt.Sprintf("--router-mode=%s", s.RouterMode))
	}
	if s.Ingress != "" {
		args = append(args, fmt.Sprintf("--ingress=%s", s.Ingress))
	}
	if s.IngressHost != "" {
		args = append(args, fmt.Sprintf("--ingress-host=%s", s.IngressHost))
	}
	if s.CreateNetworkPolicy {
		args = append(args, "--create-network-policy")
	}
	if s.EnableClusterPermissions {
		args = append(args, "--enable-cluster-permissions")
	}
	if s.RouterCPU != "" {
		args = append(args, fmt.Sprintf("--router-cpu=%s", s.RouterCPU))
	}
	if len(s.Annotations) != 0 {
		args = append(args, fmt.Sprintf("--annotations=%s", strings.Join(s.Annotations, ",")))
	}

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:        args,
					F2Namespace: s.Namespace,
				},
				Validator: &ValidateSkupperAvailable{
					Namespace:  s.Namespace,
					MaxWait:    s.MaxWait,
					SkipWait:   s.SkipStatus,
					SkipStatus: s.SkipStatus,
					Ctx:        s.Ctx,
				},
			},
		},
	}

	return phase.Run()
}

type ValidateSkupperAvailable struct {
	Namespace  *f2k8s.Namespace
	Ctx        context.Context
	MaxWait    time.Duration // If not set, defaults to types.DefaultTimeoutDuration*2
	SkipWait   bool
	SkipStatus bool

	frame2.DefaultRunDealer
	frame2.Log
}

func (v ValidateSkupperAvailable) Validate() error {
	var waitCtx context.Context
	var cancel context.CancelFunc

	wait := v.MaxWait
	if wait == 0 {
		wait = 2 * time.Minute
	}

	if !v.SkipWait {
		waitCtx, cancel = context.WithTimeout(v.Runner.OrDefaultContext(v.Ctx), wait)
		defer cancel()
	}

	phase := frame2.Phase{
		Runner: v.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Check that the router and service controller containers are reporting as ready",
				Validators: []frame2.Validator{
					&f2k8s.ContainerValidate{
						Namespace:   v.Namespace,
						PodSelector: f2sk1const.RouterSelector,
						StatusCheck: true,
					},
					&f2k8s.ContainerValidate{
						Namespace:   v.Namespace,
						PodSelector: f2sk1const.ServiceControllerSelector,
						StatusCheck: true,
					},
				},
				ValidatorRetry: frame2.RetryOptions{
					Ctx:        waitCtx,
					Ensure:     5, // The containers may briefly report ready before crashing
					KeepTrying: true,
				},
				SkipWhen: v.SkipWait,
			}, {
				Modify: &CliSkupperVersion{
					Namespace: v.Namespace,
					Ctx:       waitCtx,
				},
				SkipWhen: v.SkipStatus,
			}, {
				Validator: &Status{
					Namespace: v.Namespace,
				},
				SkipWhen: v.SkipStatus,
			},
		},
	}
	return phase.Run()
}
