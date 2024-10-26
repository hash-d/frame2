package skupperexecute

import (
	"context"
	"fmt"
	"regexp"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

// With only a namespace configured, this command will simply
// execute skupper status.  If Verbose is set, the -v flag is
// added to that execution.
//
// If the other fields are set, different verifications will
// be done on the output of the command, ensuring it matches
// the expectation.
type Status struct {
	Namespace *f2k8s.Namespace
	Ctx       context.Context

	Verbose bool

	CheckStatus bool
	Enabled     bool

	// Only TotalConn is supported for non-verbose
	CheckConnectionCounts bool
	TotalConn             int
	DirectConn            int
	IndirectConn          int

	// TODO non-zero count is still not implemented
	CheckServiceCount bool
	ExposedServices   int

	// Mode and Sitename checks are not supported for
	// non-verbose
	Mode     string
	SiteName string

	// Policies check is not yet implemetned for
	// non-verbose TODO
	CheckPolicies bool
	Policies      bool

	frame2.Log
	frame2.DefaultRunDealer
	execute.SkupperVersionerDefault
}

// TODO: replace this by f2k8s.Namespace
func (s Status) GetNamespace() string {
	return s.Namespace.GetNamespaceName()
}

// TODO: move this to a new SkupperInstallVAN or something; leave SkupperInstall as a
// SkupperOp that calls either that or CliSkupperInit
func (s Status) Execute() error {

	cmd := execute.Cmd{
		ForceOutput: true,
	}

	stdout := []string{}
	stdoutNot := []regexp.Regexp{}
	reStdout := []regexp.Regexp{}

	version := s.SkupperVersionerDefault.WhichSkupperVersion([]string{"1.4", "1.5"})
	if s.Verbose {
		if version == "1.4" {
			// TODO: in the future, return an error that means 'test skipped due to version'
			// -v was added in 1.5
			return nil
		}
		if s.CheckStatus {
			if s.Enabled {
				stdoutNot = append(stdoutNot, *regexp.MustCompile("Skupper is not enabled in namespace"))
			} else {
				stdout = append(stdout, "Skupper is not enabled in namespace")
			}
		}
		if s.CheckConnectionCounts {
			reStdout = append(reStdout, *regexp.MustCompile(fmt.Sprintf("total connections: *%d", s.TotalConn)))
			reStdout = append(reStdout, *regexp.MustCompile(fmt.Sprintf("direct connections: *%d", s.DirectConn)))
			reStdout = append(reStdout, *regexp.MustCompile(fmt.Sprintf("indirect connections: *%d", s.IndirectConn)))
		}
		if s.CheckServiceCount {
			reStdout = append(reStdout, *regexp.MustCompile(fmt.Sprintf("exposed services: *%d", s.ExposedServices)))
		}
		if s.Mode != "" {
			reStdout = append(reStdout, *regexp.MustCompile(fmt.Sprintf("mode: *%s", s.Mode)))

		}
		if s.SiteName != "" {
			reStdout = append(reStdout, *regexp.MustCompile(fmt.Sprintf("site name: *%s", s.SiteName)))

		}
		if s.CheckPolicies {
			var policyStatusStr string
			if s.Policies {
				policyStatusStr = "enabled"
			} else {
				policyStatusStr = "disabled"
			}
			reStdout = append(reStdout, *regexp.MustCompile(fmt.Sprintf("policies: *%s", policyStatusStr)))
		}
	} else {
		// For non-verbose, we use stdout (not regexp); order of checks is important
		if s.CheckStatus {
			if s.Enabled {
				stdoutNot = append(stdoutNot, *regexp.MustCompile("Skupper is not enabled in namespace"))
			} else {
				stdout = append(stdout, "Skupper is enabled for namespace", s.Namespace.GetNamespaceName())
			}
		}
		if s.Mode != "" {
			if version == "1.4" {
				stdout = append(stdout, fmt.Sprintf("in %s mode.", s.Mode))
			} else {
				// do nothing
				// TODO: in the future, return an error that means 'test skipped due to version'
			}
		}
		if s.CheckPolicies {
			if s.Policies {
				stdout = append(stdout, "(with policies)")
			} else {
				stdoutNot = append(stdoutNot, *regexp.MustCompile("(with policies)"))
			}
		}
		if s.CheckConnectionCounts {
			if s.TotalConn == 0 {
				stdout = append(stdout, "It is not connected to any other sites.")
			} else {
				stdout = append(stdout, fmt.Sprintf("It is connected to %d other site", s.TotalConn))
			}
			if s.IndirectConn > 0 {
				stdout = append(stdout, fmt.Sprintf("(%d indirectly)", s.IndirectConn))
			}
		}
		if s.DirectConn > 0 {
			return fmt.Errorf("CliStatus cannot check direct connections")
		}
		if s.CheckServiceCount {
			if s.ExposedServices == 0 {
				stdout = append(stdout, "It has no exposed services.")
			} else {
				stdout = append(stdout, fmt.Sprintf("It has %d exposed service", s.ExposedServices))
			}

		}
		if s.SiteName != "" {
			return fmt.Errorf("CliStatus cannot check site name in non-verbose")
		}
	}

	cmd.Expect = frame2.Expect{
		StdOut:      stdout,
		StdOutReNot: stdoutNot,
		StdOutRe:    reStdout,
	}

	args := []string{"status"}
	if s.Verbose {
		args = append(args, "-v")
	}

	phase := frame2.Phase{
		Runner: s.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:        args,
					F2Namespace: s.Namespace,
					Cmd:         cmd,
				},
			},
		},
	}

	return phase.Run()
}

func (s Status) Validate() error {
	return s.Execute()
}
