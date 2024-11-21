package f2skupper1

import (
	"context"
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1const"
	"regexp"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

type SkupperUpgrade struct {
	Runner       *frame2.Run
	Namespace    *f2k8s.Namespace
	ForceRestart bool
	SkipVersion  bool

	Wait time.Duration
	Ctx  context.Context

	// If true, skips checking the images against the manifest.  If
	// false and no manifest available, panic
	SkipManifest bool

	// Location of the manifest file to be used on the manifest/image
	// tag check.  If empty, check: (?)
	//
	// Before 1.5:
	// - Current dir (ie, etc package dir)
	// - Source root dir
	//
	// Since 1.5:
	// - Execute skupper version manifest to generate a manifest.json file
	ManifestFile string

	// If true, the upgrade output will be inspected, to ensure the message
	// "No update required in 'namespace'" was not shown.
	//
	// In practice, it makes it an error to try to upgrade a site that is
	// already in the right version
	CheckUpdateRequired bool

	// TODO: SkupperBinary (for multi-step upgrades)
}

func (s SkupperUpgrade) Execute() error {

	args := []string{"update"}

	if s.ForceRestart {
		args = append(args, "--force-restart")
	}

	ctx := s.Runner.OrDefaultContext(s.Ctx)
	var cancel context.CancelFunc
	var validators []frame2.Validator
	var waitMessage string
	if s.Wait == 0 {
		waitMessage = "; do not wait for pods to be up"
	} else {
		waitMessage = ", and wait for router and service-controller pods to be up"
		ctx, cancel = context.WithTimeout(s.Runner.OrDefaultContext(ctx), s.Wait)
		defer cancel()

		validators = []frame2.Validator{
			&f2k8s.Container{
				Namespace:   s.Namespace,
				PodSelector: f2sk1const.RouterSelector,
				StatusCheck: true,
			},
			&f2k8s.Container{
				Namespace:   s.Namespace,
				PodSelector: f2sk1const.ServiceControllerSelector,
				StatusCheck: true,
			},
		}
	}

	expect := frame2.Expect{}
	if s.CheckUpdateRequired {
		expect.StdOutReNot = []regexp.Regexp{*regexp.MustCompile("No update required in")}
	}

	phase := frame2.Phase{
		Runner: s.Runner,
		Doc:    "Upgrade Skupper and wait for the upgrade to be complete",
		MainSteps: []frame2.Step{
			{
				Doc: fmt.Sprintf("Upgrade skupper on namespace %q%v", s.Namespace.GetNamespaceName(), waitMessage),
				Modify: &CliSkupper{
					F2Namespace: s.Namespace,
					Args:        args,
					Cmd: f2general.Cmd{
						Ctx: ctx,
					},
				},
				Validators: validators,
				ValidatorRetry: frame2.RetryOptions{
					Allow:      60,
					Ignore:     10,
					Ensure:     5,
					KeepTrying: true,
					Ctx:        ctx,
				},
			}, {
				Doc: "Wait for deployment to match manifest",
				Validator: &ManifestMatchesDeployment{
					Namespace: s.Namespace,
				},
				SkipWhen: s.SkipManifest,
				ValidatorRetry: frame2.RetryOptions{
					Timeout:    time.Minute * 2,
					KeepTrying: true,
				},
			}, {
				Doc: "Show actual version after upgrade",
				Modify: &CliSkupperVersion{
					Namespace: s.Namespace,
				},
				SkipWhen: s.SkipVersion,
			},
		},
	}
	return phase.Run()
}
