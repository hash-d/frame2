package skupperexecute

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
)

// Runs skupper version.  By default, this command always shows
// its output (even if SKUPPER_TEST_VERBOSE_COMMANDS is not set),
// and it checks the output against the actual version (which,
// unless explicitly configured, comes from from the environment
// variable SKUPPER_TEST_VERSION or similar.
//
// This frame cannot run `skupper version manifest`.  For that,
// use SkupperManifest.
type CliSkupperVersion struct {
	Namespace *base.ClusterContext
	Ctx       context.Context

	// By default, CliSkupperVersion checks the output of
	// the version command against SKUPPER_TEST_VERSION,
	// when that variable is set
	SkipOutputCheck bool

	frame2.Log
	frame2.DefaultRunDealer
	execute.SkupperVersionerDefault
}

func (s CliSkupperVersion) Execute() error {

	cmd := execute.Cmd{
		ForceOutput: true,
	}

	if !s.SkipOutputCheck && s.GetSkupperVersion() != "" {
		cmd.Expect = cli.Expect{
			StdOut: []string{
				"client version",
				s.GetSkupperVersion(),
				// "transport version" boxes envVersion, to ensure it's not found
				// elsewhere in the string
				"transport version",
			},
		}
	}

	args := []string{"version"}

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:           args,
					ClusterContext: s.Namespace,
					Cmd:            cmd,
				},
			},
		},
	}

	return phase.Run()
}

func (s CliSkupperVersion) Validate() error {
	return s.Execute()
}