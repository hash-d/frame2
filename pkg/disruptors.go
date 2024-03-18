package frame2

type Disruptor interface {
	DisruptorEnvValue() string
}

type DisruptorConfigurer interface {
	Configure(string) error
}

// This is just a marker to indicate that the disruptor does
// not need to be listed on Run.AlwaysDisruptor on the test;
// just having it on the environment will suffice for it to
// take effect
type AlwaysDisruptor interface {
	// This is just a marker; it does nothing
	AlwaysDisruptor()
}

// Disruptors that implement the Inspector interface will
// have its Inspect() function called before each step is
// executed.
//
// The disruptor will then be able to analise whether that
// step is of interest for it or not, or even change the
// step's configuration
type Inspector interface {
	Inspect(step *Step, phase *Phase)
}

// PostMainSetupHook will be executed right after the setup
// phase completes, before the main steps.
type PostMainSetupHook interface {
	PostMainSetupHook(runner *Run) error
}

type FinalizerHook interface {

	// FinalizerHook will be executed at the end of the
	// test, before all other finalizer tasks, such as the
	// re-run of validators marked as final
	PreFinalizerHook(runner *Run) error

	// PostSubFinalizerHook will be executed after the
	// final validators are run.  It can be used, for example,
	// to reset the disruptor, so it starts a new on the
	// next sub test.  On an upgrade disruptor, for example,
	// the next cycle that has a subfinalizer needs to start
	// with the old version again.
	PostSubFinalizerHook(runner *Run) error
}

// This hook looks into the validations as all retries have been exhausted,
// and it allows a disruptor to inspect the error, or transform it
// into another error, or even clear the error, by returning null.
//
// The interface needs some work.  Only err is really required;
// runner and step may be removed or replaced in the future
type ValidationResultHook interface {
	ValidationResultHook(runner *Run, step Step, err error) error
}
