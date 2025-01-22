package disruptor

import (
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"
	"log"
	"os"

	frame2 "github.com/hash-d/frame2/pkg"
)

// On each namespace, only the first Skupper execution will use
// the old CLI (it is expected to be the init command).
//
// # All further commands use the new CLI
//
// TODO: This needs to be improved to recognize init commands vs
// everything else.  For that, however, SkupperCliPathSetter and
// SkupperVersioner need to be on the same frame, and they are not,
// today.  The PathSetter is on the Skuppercmd only, and the Versioner
// is on several different frames.
type AlternateSkupper struct {

	// Once this is set to true, the alternate version will be used
	// for executing Skupper commands (ie, the other version, not the
	// one used for the install)
	useAlternate map[string]bool

	// By default AlternateSkupper installs the old version and then
	// uses the new version for all other commands.  Setting this
	// option to true inverses that behavior
	NewOnInstall bool
}

func (a AlternateSkupper) DisruptorEnvValue() string {
	return "ALTERNATE_SKUPPER"
}

func (a AlternateSkupper) PreFinalizerHook(runner *frame2.Run) error {
	return nil
}

func (a *AlternateSkupper) Configure(conf string) error {
	switch conf {
	case "NEW":
		a.NewOnInstall = true
	case "", "OLD":
		a.NewOnInstall = false
	default:
		return fmt.Errorf("%q is not a valid configuration for AlternateSkupper", conf)
	}
	return nil
}

func (a *AlternateSkupper) PostSubFinalizerHook(runner *frame2.Run) error {
	log.Printf("Alternate Skupper resetting its map post sub finalizers")
	a.useAlternate = map[string]bool{}
	return nil
}

func (a *AlternateSkupper) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if a.useAlternate == nil {
		a.useAlternate = map[string]bool{}
	}
	var useAlternate bool
	_ = step.IterFrames(func(frame any) (any, error) {
		if frame, ok := frame.(f2skupper1.SkupperCliPathSetter); ok {
			ns := frame.GetNamespace()
			if ns == "" {
				useAlternate = true
			} else if a.useAlternate[ns] {
				useAlternate = true
			} else {
				useAlternate = false
			}

			useNew := a.NewOnInstall != useAlternate
			if useNew {
				log.Printf("AlternateSkupper disruptor using 'new' path for %T", frame)
				setCliPathCurrentEnv(frame)
			} else {
				log.Printf("AlternateSkupper disruptor using 'old' path for %T", frame)
				setCliPathOldEnv(frame)
			}
			a.useAlternate[ns] = true
		}
		if frame, ok := frame.(f2skupper1.SkupperVersioner); ok {
			ns := frame.GetNamespace()
			if ns == "" {
				useAlternate = true
			} else if a.useAlternate[ns] {
				useAlternate = true
			} else {
				useAlternate = false
			}

			useNew := a.NewOnInstall != useAlternate
			if useNew {
				version := os.Getenv(frame2.ENV_VERSION)
				log.Printf("AlternateSkupper disruptor resetting version to %q for %T", version, frame)
				frame.SetSkupperVersion(version)
			} else {
				version := os.Getenv(frame2.ENV_OLD_VERSION)
				log.Printf("AlternateSkupper disruptor updating version to %q for %T", version, frame)
				frame.SetSkupperVersion(version)
			}
		}
		/*
			if frame, ok := frame.(execute.SkupperUpgradable); ok {
				ns := frame.GetNamespace()
				a.useAlternate[ns] = true
			}
		*/
		return frame, nil
	})
}
