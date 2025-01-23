package disruptor

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hash-d/frame2/pkg/frames/f2skupper1"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

type TestUpgradeStrategy string

// Upgrade strategies accepted by frame2.ENV_UPGRADE_STRATEGY
const (
	UPGRADE_STRATEGY_CREATION TestUpgradeStrategy = "CREATION"

	// This one is special; it is set after a colon and inverts the
	// result. For example: ":INVERSE" or "CREATION:INVERSE"
	UPGRADE_STRATEGY_INVERSE TestUpgradeStrategy = "INVERSE"

	// Do all public first, then all private.  Within the groups,
	// they'll be left in their original order.
	UPGRADE_STRATEGY_PUB_FIRST TestUpgradeStrategy = "PUB_FIRST"

	// Do all public first, then all private.  Within the groups,
	// they'll be left in their original order.
	UPGRADE_STRATEGY_PRV_FIRST TestUpgradeStrategy = "PRV_FIRST"

	// For the SKUPPER_TEST_OLD.*IMAGE variables, if set to this value,
	// the referred variable will be set to an empty value.  This may
	// be required on environments where trying to set a variable to
	// the empty value actually unsets it (such as Jenkins' withEnv)
	VAR_EMPTY_VALUE = "frame2.empty"
)

// Returns the Upgrade strategy configured in the environment
func getUpgradeStrategy() (TestUpgradeStrategy, bool) {
	var invert bool
	var strategy TestUpgradeStrategy

	envValue := os.Getenv(frame2.ENV_UPGRADE_STRATEGY)

	s := strings.SplitN(envValue, ":", 2)
	strategy = TestUpgradeStrategy(s[0])
	if strategy == "" {
		strategy = UPGRADE_STRATEGY_CREATION
	}
	if len(s) > 1 {
		if s[1] == string(UPGRADE_STRATEGY_INVERSE) {
			invert = true
		} else {
			panic(fmt.Sprintf("invalid option to SKUPPER_TEST_UPGRADE_STRATEGY: %v", s[1]))
		}
	}

	return strategy, invert

}

// Return the public and private contexts in different slices, but keeping
// their relative orders.
func getPubPrvUpgradeTargets(targets []*f2k8s.Namespace) (pubs, privs []*f2k8s.Namespace) {
	for _, c := range targets {
		if c.GetKind() == f2k8s.Private {
			privs = append(privs, c)
		} else {
			pubs = append(pubs, c)
		}
	}
	return pubs, privs
}

// Sort the targets according to some strategy, configured on
// SKUPPER_TEST_UPGRADE_STRATEGY.  If none set, return the target list
// unchanged
func sortUpgradeTargets(targets []*f2k8s.Namespace) []*f2k8s.Namespace {

	var ret []*f2k8s.Namespace

	strategy, invert := getUpgradeStrategy()

	switch strategy {
	case UPGRADE_STRATEGY_CREATION:
		ret = targets[:]
	case UPGRADE_STRATEGY_PUB_FIRST:
		pubs, privs := getPubPrvUpgradeTargets(targets)
		ret = append(pubs, privs...)
	case UPGRADE_STRATEGY_PRV_FIRST:
		pubs, privs := getPubPrvUpgradeTargets(targets)
		ret = append(privs, pubs...)
	default:
		panic(fmt.Sprintf("invalid upgrade strategy: %v", strategy))
	}

	if invert {
		lenRet := len(ret)
		for i := 0; i < lenRet/2; i++ {
			ret[i], ret[lenRet-i-1] = ret[lenRet-i-1], ret[i]
		}
	}

	return ret
}

func upgradeSites(targets []*f2k8s.Namespace, runner *frame2.Run) error {
	var steps []frame2.Step

	log.Printf("Upgrading sites %+v", targets)

	for _, t := range targets {
		steps = append(steps, frame2.Step{
			Doc: "Upgrade Skupper",
			Modify: f2skupper1.SkupperUpgrade{
				Runner:    runner,
				Namespace: t,
				Wait:      time.Minute * 10,
				// Perhaps this should be configurable.  In any case,
				// if we're trying to upgrade something that was already
				// upgraded, we probably have something wrong on the test
				// or on the frames.
				CheckUpdateRequired: true,
			},
		})
	}
	phase := frame2.Phase{
		Runner:    runner,
		MainSteps: steps,
		Doc:       "Upgrade sites per disruptor",
	}
	return phase.Run()
}

// At the end of the test, before the tear down, upgrade all
// sites and then re-run all tests marked as final
//
// This is a very basic upgrade test; it's cheap and simple
//
// The upgrade strategy can be defined on the environment
// variable SKUPPER_TEST_UPGRADE_STRATEGY.
//
// TODO: if at the finalize step, there are no final validators,
// the test should be marked as a failure (as it did not deliver
// what the disruptor promised)
type UpgradeAndFinalize struct {
	targets []*f2k8s.Namespace
	useNew  bool
}

func (u UpgradeAndFinalize) DisruptorEnvValue() string {
	return "UPGRADE_AND_FINALIZE"
}

func (u *UpgradeAndFinalize) PreFinalizerHook(runner *frame2.Run) error {
	var steps []frame2.Step
	u.useNew = true

	targets := sortUpgradeTargets(u.targets)

	for _, t := range targets {
		steps = append(steps, frame2.Step{
			Doc: fmt.Sprintf("Disruptor UpgradeAndFinalize: upgrade namespace %v", t.GetNamespaceName()),
			Modify: &f2skupper1.SkupperUpgrade{
				Runner:    runner,
				Namespace: t,
				Wait:      time.Minute * 10,
			},
		})
	}
	phase := frame2.Phase{
		Runner:    runner,
		Doc:       "Disruptor UpgradeAndFinalize: Upgrade phase",
		MainSteps: steps,
	}
	return phase.Run()
}

func (u *UpgradeAndFinalize) PostSubFinalizerHook(runner *frame2.Run) error {
	log.Printf("UpgradeAndFinalize: clearing targets on PostSubFinalizerHook")
	u.useNew = false
	u.targets = []*f2k8s.Namespace{}
	return nil
}

func (u *UpgradeAndFinalize) Inspect(step *frame2.Step, phase *frame2.Phase) {
	err := step.IterFrames(func(frame any) (any, error) {
		if frame, ok := frame.(f2skupper1.SkupperUpgradable); ok {
			u.targets = append(u.targets, frame.SkupperUpgradable())
		}
		if frame, ok := frame.(f2skupper1.SkupperCliPathSetter); ok {
			if !u.useNew {
				log.Printf("UpgradeAndFinalize disruptor updating path for %T", frame)
				setCliPathOldEnv(frame)
			} else {
				log.Printf("UpgradeAndFinalize disruptor resetting path for %T", frame)
				setCliPathCurrentEnv(frame)
			}
		}
		if frame, ok := frame.(f2skupper1.SkupperVersioner); ok {
			if !u.useNew {
				version := os.Getenv(frame2.ENV_OLD_VERSION)
				log.Printf("UpgradeAndFinalize disruptor updating version to %q for %T", version, frame)
				frame.SetSkupperVersion(version)
			} else {
				version := os.Getenv(frame2.ENV_VERSION)
				log.Printf("UpgradeAndFinalize disruptor resetting version to %q for %T", version, frame)
				frame.SetSkupperVersion(version)
			}
		}

		return frame, nil
	})
	if err != nil {
		panic(fmt.Sprintf("Add error return to Inspect: %v", err))
	}
}

// Undoes any changes done by setCliPathOldEnv, restoring the action to its original
// setting.
//
// TODO: This does not cover the case where the the path was already set before
// setCliPathOldEnv() was called (so that value is not restored).  This simply
// sets the path to "" and the environment to []string{}, so the action uses the
// actual environment variables and skupper binary on the path.
func setCliPathCurrentEnv(action f2skupper1.SkupperCliPathSetter) {
	// For those SKUPPER_TEST_OLD image variables that are set, we change them
	// on the environment for the called command
	var env []string
	log.Printf("Action %T restored with empty environment and skupper from PATH", action)
	action.SetSkupperCliPath("", env)
}

// Sets the path to the Skupper executable on this action to the one set on
// SKUPPER_TEST_OLD_BIN, and sets the execution environment to add or overwrite
// any Skupper IMAGE variables with their SKUPPER_TEST_OLD settings
func setCliPathOldEnv(action f2skupper1.SkupperCliPathSetter) {
	path := os.Getenv(frame2.ENV_OLD_BIN)
	if path == "" {
		panic("Upgrade disruptor requested, but no SKUPPER_TEST_OLD_BIN config")
	}

	// For those SKUPPER_TEST_OLD image variables that are set, we change them
	// on the environment for the called command
	var env []string
	for oldEnvKey, envKey := range frame2.EnvOldMap {
		// Do not change to os.GetEnv: we want the ability to unset a variable
		// for the old version
		if image, ok := os.LookupEnv(oldEnvKey); ok {
			if image == VAR_EMPTY_VALUE {
				image = ""
			}
			env = append(env, fmt.Sprintf("%s=%s", envKey, image))
		}

	}

	log.Printf(
		"Action %T updated with path %q and additional environment %+v",
		action,
		path,
		env,
	)

	action.SetSkupperCliPath(path, env)
}

// Right after setup is complete, update part of the VAN, and
// then run the tests in this mixed-version network
//
// At the end of the test, before the tear down, upgrade the
// remaining sites and then re-run all tests marked as final
//
// The upgrade strategy can be defined on the environment
// variable SKUPPER_TEST_UPGRADE_STRATEGY.
//
// When using a strategy such as PUB_FIRST, the public sites
// will be done on the postSetup hook, and the others in the
// PreFinalizer.  On other strategies, the list will simply
// be split in two halves
type MixedVersionVan struct {
	targets   []*f2k8s.Namespace
	remaining []*f2k8s.Namespace
	useNew    bool
}

func (m MixedVersionVan) DisruptorEnvValue() string {
	return "MIXED_VERSION_VAN"
}

func (m *MixedVersionVan) PostMainSetupHook(runner *frame2.Run) error {
	m.useNew = true
	targets := sortUpgradeTargets(m.targets)

	var cycleTargets, nextCycle []*f2k8s.Namespace

	strategy, _ := getUpgradeStrategy()

	switch strategy {
	default:
		cycleTargets = targets[:len(targets)/2]
		nextCycle = targets[len(targets)/2:]
	}

	m.remaining = nextCycle

	return upgradeSites(cycleTargets, runner)
}

// Updates the remaining sites before the finalizer runs
func (m *MixedVersionVan) PreFinalizerHook(runner *frame2.Run) error {
	m.useNew = true

	targets := sortUpgradeTargets(m.remaining)
	m.remaining = []*f2k8s.Namespace{}

	return upgradeSites(targets, runner)
}

func (u *MixedVersionVan) PostSubFinalizerHook(runner *frame2.Run) error {
	u.useNew = false
	u.targets = []*f2k8s.Namespace{}
	return nil
}

// Change this to a mix-in, share with UpgradeAndFinalize?
func (m *MixedVersionVan) Inspect(step *frame2.Step, phase *frame2.Phase) {
	err := step.IterFrames(func(frame any) (any, error) {
		if UpgradableStep, ok := frame.(f2skupper1.SkupperUpgradable); ok {
			m.targets = append(m.targets, UpgradableStep.SkupperUpgradable())
		}
		if pathSetAction, ok := frame.(f2skupper1.SkupperCliPathSetter); ok {
			if !m.useNew {
				log.Printf("MixedVersionVan disruptor updating path on %T", pathSetAction)
				setCliPathOldEnv(pathSetAction)
			} else {
				log.Printf("MixedVersionVan disruptor resetting path on %T", pathSetAction)
				setCliPathCurrentEnv(pathSetAction)
			}
		}
		if action, ok := frame.(f2skupper1.SkupperVersioner); ok {
			if !m.useNew {
				version := os.Getenv(frame2.ENV_OLD_VERSION)
				log.Printf("MixedVersionVan disruptor updating version to %q for %T", version, action)
				action.SetSkupperVersion(version)
			} else {
				version := os.Getenv(frame2.ENV_VERSION)
				log.Printf("MixedVersionVan disruptor resetting version to %q for %T", version, action)
				action.SetSkupperVersion(version)
			}
		}
		return frame, nil
	})
	if err != nil {
		panic(fmt.Sprintf("Add error return to Inspect: %v", err))
	}
}
