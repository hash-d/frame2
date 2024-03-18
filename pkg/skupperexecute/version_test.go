package skupperexecute_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/disruptors"
	"github.com/hash-d/frame2/pkg/environment"
	"github.com/hash-d/frame2/pkg/skupperexecute"
	"gotest.tools/assert"
)

// This is not much of a test, as the frame itself does not do much.  When it is
// run, though, the environment may make it more interesting;
//
// - SKUPPER_TEST_VERSION being set will activate its version checking mechanism
// - Some Upgrade disruptor may do the same, and for both old and new versions
//
// So, it's still valid running it for full tests, especially when disrupted.
//
// This frame is also run as part of Install and Upgrade frames, so it is also
// partially tested on those tests.
func TestVersion(t *testing.T) {

	r := &frame2.Run{
		T: t,
	}
	defer r.Finalize()

	r.AllowDisruptors([]frame2.Disruptor{
		&disruptors.UpgradeAndFinalize{},
	})

	envSetup := &environment.JustSkupperDefault{
		Name:         "version-test",
		AutoTearDown: true,
	}

	setup := frame2.Phase{
		Runner: r,
		Doc:    "Create a skupper installation",
		Setup: []frame2.Step{
			{
				Modify: envSetup},
		},
	}
	assert.Assert(t, setup.Run())

	// We don't care whether the test is using pub or prv, here;
	// we just pick the first environment it has
	ns := envSetup.Topo.ListAll()[0]

	test := frame2.Phase{
		Runner: r,
		Doc:    "Main test phase",
		MainSteps: []frame2.Step{
			{
				Doc: "Execute skupper version",
				// TODO Some changes needed here...  First, we
				// should be using a SkupperOp named just
				// f2skupper.Version.  Second, that should be a
				// Validator, not an Executor, and then we could
				// mark it as a final validator directly.
				Validator: &skupperexecute.CliSkupperVersion{
					Namespace: ns,
				},
				ValidatorFinal: true,
			},
		},
	}

	assert.Assert(t, test.Execute())

}
