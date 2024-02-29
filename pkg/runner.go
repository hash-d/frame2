package frame2

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/skupperproject/skupper/test/utils/base"
)

// Each step on a phase runs with its own Runner.  These constants identify
// what type of work is being done by the Runner.  The step IDs are derived
// from their Runner type
type RunnerType int

const (
	RootRunner RunnerType = iota
	PhaseRunner
	ValidatorRunner
	ModifyRunner
	SetupRunner
	HookRunner
	SubTestRunner
	StepRunner
	TearDownRunner
	MonitorRunner
)

// The Run (TODO rename?) keeps context accross the execution of the test.  Each
// phase and each step has its runner, in a tree structure
type Run struct {
	T                  *testing.T
	Doc                string      // TODO this is currently unused (ie, it's not printed)
	RequiredDisruptors []Disruptor // TODO
	savedT             *testing.T  // TODO: review.  Only private + getter/setter?
	monitors           []*Monitor
	finalValidators    []Validator
	subFinalValidators []Validator
	ctx                context.Context
	cancelCtx          context.CancelFunc
	root               *Run // TODO Perhaps replace this by a recursive call, now we have parent
	disruptor          []Disruptor
	parent             *Run
	children           []*Run
	kind               RunnerType
	sequence           int // RunId is derived from kind+sequence
	nextChildSequence  int
	postSetup          bool
	postMainSetupDone  bool
	named              bool
}

// Return the full ID of the Runner, which includes the ID of its parent
func (r *Run) GetId() string {
	if r == nil {
		return "-"
	}
	var kindLetter string
	switch r.kind {
	case RootRunner:
		kindLetter = "R"
	case PhaseRunner:
		kindLetter = "p"
	case ValidatorRunner:
		kindLetter = "v"
	case ModifyRunner:
		kindLetter = "m"
	case SetupRunner:
		kindLetter = "set"
	case HookRunner:
		kindLetter = "H"
	case StepRunner:
		kindLetter = "s"
	case SubTestRunner:
		kindLetter = "SubT"
	case TearDownRunner:
		kindLetter = "TD"
	case MonitorRunner:
		kindLetter = "M"
	default:
		panic("unhandled kind of Runner")
	}
	localId := fmt.Sprintf("%v%v", kindLetter, r.sequence)
	if r.parent == nil || r.kind == SubTestRunner {
		return fmt.Sprintf("%v", localId)
	}
	return fmt.Sprintf("%v.%v", r.parent.GetId(), localId)
}

// TODO: make just Child(), which reuses the runner's own T
func (r *Run) ChildWithT(t *testing.T, kind RunnerType) *Run {
	// TODO Should we allow this, or panic?
	if r == nil {
		return nil
	}
	root := r.root
	if root == nil {
		root = r
	}
	child := Run{
		parent:    r,
		T:         t,
		savedT:    t,
		ctx:       r.ctx,
		cancelCtx: r.cancelCtx,
		sequence:  r.nextChildSequence,
		root:      root,
		kind:      kind,
	}
	r.nextChildSequence += 1
	r.children = append(r.children, &child)

	return &child
}

func (r *Run) ReportChildren(indent int) {
	for _, c := range r.children {
		log.Printf("%v- %v %+v", strings.Repeat(" ", indent), c.GetId(), *c)
		c.ReportChildren(indent + 1)
	}
}

// GetContext will always return a context.  If not defined on the
// current level, check the parent.  If not on the parent, create a
// new Background context.
//
// When contexts are created on this method, they get scheduled for
// cancellation on T.Cleanup()
//
// TODO contexts are not expected to change?
func (r *Run) GetContext() context.Context {
	if r == nil {
		return context.Background()
	}
	if r.ctx != nil {
		return r.ctx
	}
	if ctx := r.parent.GetContext(); ctx != nil {
		return ctx
	}
	r.ctx, r.cancelCtx = context.WithCancel(context.Background())
	r.savedT.Cleanup(r.cancelCtx)
	return r.ctx
}

// Return ctx if not nil.  If nil, return the runner's default context
//
// # If the runner is not available (call on nil reference), return context.Background
//
// TODO review
func (r *Run) OrDefaultContext(ctx context.Context) context.Context {
	if r == nil {
		return context.Background()
	}
	if ctx == nil {
		return r.GetContext()
	}
	return ctx
}

func (r *Run) CancelContext() {
	r.cancelCtx()
}

func (r *Run) addMonitor(step *Monitor) {

	r.monitors = append(r.monitors, step)
}

func (r *Run) addFinalValidators(v []Validator) {
	root := r.getRoot()
	root.finalValidators = append(root.finalValidators, v...)
}

func (r *Run) addSubFinalValidators(v []Validator) {
	namedTest := r.getNamed()
	namedTest.subFinalValidators = append(namedTest.subFinalValidators, v...)
}

func (r *Run) getNamed() *Run {
	switch {
	case r.named:
		return r
	case r.parent == nil:
		return r
	case r == r.getRoot():
		return r
	default:
		return r.parent.getNamed()
	}
}

func (r *Run) getRoot() *Run {
	//	if r == nil {
	//		return nil
	//	}
	if r.root == nil {
		return r
	} else {
		return r.root
	}
}

func (r *Run) getDisruptors() []Disruptor {
	return append(r.getRoot().disruptor, r.getRoot().RequiredDisruptors...)
}

// Run steps that are still part of the subtest, but must be run at its very end,
// right before the tear down.  Failures here will count as test failure
func (r *Run) subFinalize() {
	// TODO Today, it runs the pre-subfinalizer-hook only when there are listed
	// subfinalizers.  Perhaps that's ok, but check if there are situations where
	// the hook needs to be run without subfinalizers.  It's like that today to
	// avoid clashes between main test disruptors and subtest disruptors
	if len(r.subFinalValidators) > 0 {
		for _, d := range r.getDisruptors() {
			if d, ok := d.(PreFinalizerHook); ok {
				log.Printf("[R] Running pre-subfinalizer hook")
				var err error
				r.T.Run("pre-subfinalizer-hook", func(t *testing.T) {
					err = d.PreFinalizerHook(r.ChildWithT(t, HookRunner))
				})
				if err != nil {
					r.T.Errorf("pre-subfinalizer hook failed: %v", err)
				}
			}
		}
		log.Printf("[R] Running sub test finalizers")

		// TODO Simplify: instead of t.Run() + phase, use just one named phase.
		r.T.Run("subfinal-validator-re-run", func(t *testing.T) {
			log.Printf("[R] Running sub test final validators")
			subPhase := Phase{
				Runner: r.ChildWithT(t, ValidatorRunner),
				MainSteps: []Step{
					{
						Validators: r.subFinalValidators,
						ValidatorRetry: RetryOptions{
							Allow: base.GetEnvInt(ENV_FINAL_RETRY, 1),
						},
					},
				},
			}

			subPhase.Run()

		})
	}
}

// Run steps that are still part of the test, but must be run at its very end,
// right before the tear down.  Failures here will count as test failure
func (r *Run) Finalize() {
	for _, d := range r.getDisruptors() {
		if d, ok := d.(PreFinalizerHook); ok {
			log.Printf("[R] Running pre-finalizer hook")
			var err error
			r.T.Run("pre-finalizer-hook", func(t *testing.T) {
				err = d.PreFinalizerHook(r.ChildWithT(t, HookRunner))
			})
			if err != nil {
				r.T.Errorf("pre-finalizer hook failed: %v", err)
			}
		}
	}
	log.Printf("[R] Running finalizers")

	if len(r.finalValidators) > 0 {
		r.T.Run("final-validator-re-run", func(t *testing.T) {
			log.Printf("[R] Running final validators")
			// TODO: change this by a phase run with Retry and a slice of
			// validators, taking care of the runner.  Perhaps make a copy of
			// the validators, instead of using pointers?
			fn := func() error {
				failed := false
				var err, last_err error
				for _, v := range r.finalValidators {
					err = v.Validate()
					if err != nil {
						failed = true
						last_err = err
					}
				}
				if failed {
					return fmt.Errorf("at least one final validator failed.  Last err: %v", last_err)
				}
				return nil
			}
			_, err := Retry{
				Fn: fn,
				Options: RetryOptions{
					Allow: base.GetEnvInt(ENV_FINAL_RETRY, 1),
				},
			}.Run()
			if err != nil {
				log.Printf("final validation failed: %v", err)
				// This dummy step is required while logic above not changed
				// to use a phase?
				dummyStep := Step{
					Name: "Final validator run",
				}
				if err = validationResultHook(r, dummyStep, err); err != nil {
					t.Errorf("final validation failed: %v", err)
				}

			}
		})
	}
	// TODO add some if debug
	// r.ReportChildren(0)
}

// This will cause all active monitors to report their status on the logs.
//
// It should generally be run as defer r.Report(), right after the Run creation
func (r *Run) Report() {

	failed := false
	for _, m := range r.monitors {
		err := (*m).Report()
		if err != nil {
			failed = true
		}
	}
	if failed {
		r.savedT.Errorf("At least one monitor failed")
	}

}

// List the disruptors that a test accepts, and initialize a disruptor if
// SKUPPER_TEST_DISRUPTOR set on the environment matches a disruptor from the list.
//
// If any of the values on the environment variable does not match a value on
// the list, the test will be skipped in this run (ie, a disruptor test was
// requested, but the test does not allow for it).
//
// If the environment variable is empty, this is a no-op.
//
// Attention when calling with disruptors that use pointer reference methods: define
// them on the list as a reference to the struct.  Otherwise, the pointer reference
// methods will not be part of the method set, and some interfaces may not match
func (r *Run) AllowDisruptors(list []Disruptor) {
	kind := os.Getenv("SKUPPER_TEST_DISRUPTOR")

	if kind == "" {
		// No disruptor requested
		return
	}

	if r.getRoot().disruptor != nil {
		r.T.Fatalf("attempt to re-define the disruptor. Was %s", r.getRoot().disruptor)
	}

	disruptor_list := strings.Split(kind, ";")

outer:
	for _, d := range disruptor_list {
		name, conf, _ := strings.Cut(d, ":")
		for _, allowed := range list {
			if name == allowed.DisruptorEnvValue() {
				log.Printf("DISRUPTOR: %v", name)
				if conf != "" {
					if allowed, ok := allowed.(DisruptorConfigurer); ok {
						err := allowed.Configure(conf)
						if err != nil {
							panic(fmt.Sprintf("Failed configuration for %q: %v", name, err))
						}
					} else {
						panic(fmt.Sprintf("Disruptor %q does not accept configuration", name))
					}
					log.Printf("Configured disruptor: %+v", allowed)
				}
				r.getRoot().disruptor = append(r.getRoot().disruptor, allowed)
				continue outer
			}
		}
		r.T.Skipf("This test does not support the disruptor %q", d)
	}

	/* TODO This should replace the loop above; check for any AlwaysDisruptors
	 *      However, how to get the Disruptor class for the given env variable
	 *      value?  This will need some type of plugin 'subscription'
	for _, d := range r.getRoot().disruptor {
		if _, ok := d.(AlwaysDisruptor); ok {
			log.Printf("DISRUPTOR: %v", d)
			r.getRoot
		}
	}
	*/

}

// While the Run keeps context accross the test, the phases do the actual work,
// and they allow the test to be split in multiple pieces.
//
// This allows to:
// - Put a phase on a loop
// - Access variables populated in previous phases
type Phase struct {
	Name      string
	Doc       string
	Setup     []Step
	Teardown  []Step
	MainSteps []Step
	//BaseRunner *base.ClusterTestRunnerBase
	teardowns []Executor
	Runner    *Run

	savedRunner *Run
	previousRun bool
	connected   bool

	Log
	DefaultRunDealer
}

// TODO Review/remove
func (p *Phase) GetRunner() *Run {
	if r := p.DefaultRunDealer.GetRunner(); r != nil {
		return r
	}
	if r := p.Runner; r != nil {
		return r
	}

	return nil
}

// TODO: move this into Phase?
// Checks whether a step should be skipped, and whether it should be treated
// as an individual subtest; the heavy lifting is done on processStep_
func processStep(t *testing.T, step Step, Log FrameLogger, p *Phase, kind RunnerType) error {
	// TODO: replace [R] with own logger with Prefix?
	var err error

	id := p.DefaultRunDealer.GetRunner().GetId()

	if step.SkipWhen {
		Log.Printf("[R] %v step with modifier %T skipped (%s)", id, step.Modify, step.Doc)
		return nil
	}

	if step.Name != "" {
		// For a named test, run or fail, we work the same.  It's up to t to
		// mark it as failed
		Log.Printf("Entering subtest with id %v", id)
		_ = t.Run(step.Name, func(t *testing.T) {
			//log.Printf("[R] %v current test: %q", id, t.Name())
			Log.Printf("[R] %v Doc: %v", id, step.Doc)
			processErr := processStep_(t, step, SubTestRunner, Log, p, true)
			if processErr != nil {
				// This makes it easier to find the failures in log files
				Log.Printf("[R] test %v - %q failed", id, t.Name())
				// For named tests, we do not return the error up; we
				// just mark it as a failed test
				t.Errorf("test failed: %v", processErr)
			}
			Log.Printf("[R] %v step Subtest %q completed", id, t.Name())
		})
	} else {
		/*
			name := "-"
			if t != nil {
				name = t.Name()
			}
			Log.Printf("[R] %v current test: %q", id, name)
		*/
		err = processStep_(t, step, kind, Log, p, false)
		//Log.Printf("[R] Step %q result %+v", id, err)
	}
	return err

}

// Does the heavy lifting of executing a single step from a phase; execute each of its
// parts: setup, modify, substeps, validations, etc
func processStep_(t *testing.T, step Step, kind RunnerType, Log FrameLogger, p *Phase, named bool) error {

	stepRunner := p.DefaultRunDealer.GetRunner().ChildWithT(t, kind)
	stepRunner.named = named
	if named {
		defer stepRunner.subFinalize()
	}
	id := stepRunner.GetId()
	Log.Printf("[R] %v doc %q", id, step.Doc)

	for _, disruptor := range p.GetRunner().getDisruptors() {
		if disruptor != nil {
			if disruptor, ok := disruptor.(Inspector); ok {
				disruptor.Inspect(&step, p)
			}
		}
	}

	if step.Modify != nil {
		var modifyRunner *Run
		if mod, ok := step.Modify.(RunDealer); ok {
			mod.SetRunner(stepRunner, ModifyRunner)
			modifyRunner = mod.GetRunner()
		} else {
			modifyRunner = stepRunner.ChildWithT(t, ModifyRunner)
		}
		id := modifyRunner.GetId()
		Log.Printf("[R] %v Modifier %T", id, step.Modify)
		var err error
		start := time.Now()
		if phase, ok := step.Modify.(Phase); ok {
			err = phase.runP(modifyRunner)
		} else {
			// This is a simple executor; we just execute it
			err = step.Modify.Execute()
		}
		duration := time.Now().Sub(start)
		if err != nil {
			Log.Printf("[R] %v modify-not-ok %T (%v)", id, step.Modify, duration)
			return fmt.Errorf("modify step failed: %w", err)
		} else {
			Log.Printf("[R] %v modify-ok %T (%v)", id, step.Modify, duration)
		}
	}

	subStepList := step.Substeps
	if step.Substep != nil {
		subStepList = append([]*Step{step.Substep.GetStep()}, step.Substeps...)
	}
	for _, subStep := range subStepList {
		_, err := Retry{
			Fn: func() error {
				return processStep(t, *subStep, Log, p, SubTestRunner)
			},
			Options: step.SubstepRetry,
		}.Run()
		if err != nil {
			return fmt.Errorf("substep failed: %w", err)
		}

	}

	validatorList := step.Validators
	if step.Validator != nil {
		validatorList = append([]Validator{step.Validator}, validatorList...)
	}

	if len(validatorList) > 0 {
		start := time.Now()
		for _, v := range validatorList {
			if val, ok := v.(RunDealer); ok {
				val.SetRunner(stepRunner, ValidatorRunner)
			}
		}

		// This is a generic Runner, if the validtor is not a RunDealer
		// TODO remove this once all actions are RunDealers
		validatorRunner := stepRunner.ChildWithT(t, ValidatorRunner)
		if step.ValidatorFinal {
			p.savedRunner.addFinalValidators(validatorList)
		}
		if step.ValidatorSubFinal {
			p.savedRunner.addSubFinalValidators(validatorList)
		}
		fn := func() error {
			someFailure := false
			someSuccess := false
			var lastErr error
			var lastErrValidator Validator
			for i, v := range validatorList {
				var id string
				if v, ok := v.(RunDealer); ok {
					id = v.GetRunner().GetId()
				} else {
					id = validatorRunner.GetId()
				}

				// TODO remove this once the SetRunner thing is fixed above
				//      (see the TODO TODO TODO line)
				//id = validatorRunner.GetId()

				Log.Printf("[R] %v.v%d Validator %T", id, i, v)
				// TODO: create and set individual runners for each validator?
				err := v.Validate()
				if err == nil {
					someSuccess = true
				} else {
					someFailure = true
					lastErr = err
					lastErrValidator = v
					Log.Printf("[R] %v.v%d Validator %T failed: %v", id, i, v, err)
					if step.ExpectError {
						Log.Printf("[R] (error expected)")
						lastErr = nil
					}
					// Error or not, we do not break or return; we check all
				}
			}
			if step.ExpectError && someSuccess {
				return fmt.Errorf("error expected, but at least one validator passed")
			}
			if !step.ExpectError && someFailure {
				return fmt.Errorf("at least one validator failed.  last error (on %T): %w", lastErrValidator, lastErr)
			}
			return lastErr
		}

		_, err := Retry{
			Fn:      fn,
			Options: step.ValidatorRetry,
		}.Run()
		elapsed := time.Now().Sub(start)
		if err == nil {
			Log.Printf("[R] %v validation-ok (%v)", id, elapsed)
		} else {
			Log.Printf("[R] %v validation-not-ok: %v (%v)", id, err, elapsed)
		}
		return validationResultHook(p.GetRunner(), step, err)
	}
	return nil
}

// Hook for validation result; the handler may change the err (wrap,
// turn into nil or even change it to some other error altogether).
func validationResultHook(runner *Run, step Step, err error) error {
	for _, d := range runner.getDisruptors() {
		if d, ok := d.(ValidationResultHook); ok {
			err = d.ValidationResultHook(runner, step, err)
		}
	}
	return err
}

// For a Phase that did not define a Run, this will create a Run
// and set its T accordingly
//
// This is only for the simplest case, when a single phase is
// required.
//
// If the Phase already had a Runner set, it will fail.
func (p *Phase) RunT(t *testing.T) error {
	if p.GetRunner() == nil {
		p.Runner = &Run{
			T: t,
		}
		p.savedRunner = p.Runner
		p.DefaultRunDealer.Runner = p.Runner
	} else {
		return fmt.Errorf("Phase.RunT configuration error: cannot reset the Runner")
	}
	return p.Run()
}

func (p Phase) Execute() error {
	return p.Run()
}

func (p *Phase) Run() error {
	runner := p.GetRunner()
	if runner == nil {
		fmt.Printf("##################################\n")
		fmt.Printf("Nil runner on phase %q/%q\n", p.Name, p.Doc)
		fmt.Printf("%#v\n", p)
		debug.PrintStack()
		fmt.Printf("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n")
		runner = p.Runner
	}
	return p.runP(runner)
}

// Run phase; it creates a child runner for the given one
func (p *Phase) runP(runner *Run) error {
	var err error

	var id string

	// If a named phase, and within a *testing.T, create a subtest
	if p.Name != "" && p.GetRunner().T != nil {
		ok := p.GetRunner().T.Run(p.Name, func(t *testing.T) {
			p.DefaultRunDealer.Runner = runner.ChildWithT(t, PhaseRunner)
			p.Runner.named = true
			id = p.GetRunner().GetId()
			log.Printf("[R] %v current test: %q", id, t.Name())
			p.Log.Printf("[R] %v Phase doc: %v", id, p.Doc)
			err = p.run()
			p.GetRunner().subFinalize()
			p.Log.Printf("[R] %v phase Subtest %q completed", id, t.Name())
		})

		//p.Runner = savedRunner
		if !ok && err == nil {
			err = fmt.Errorf("phase %q returned not-ok, but no errors", p.Name)
		}
	} else {
		// otherwise, just run it
		//log.Printf("[R] %vcurrent test: %q", id, p.Runner.T.Name())
		p.SetRunner(runner, PhaseRunner)
		id = p.GetRunner().GetId()
		p.Log.Printf("[R] %v Phase doc: %v", id, p.Doc)
		err = p.run()
	}

	if err != nil {
		if p.GetRunner().T == nil {
			p.Log.Printf("[R] %v Phase error: %v", id, err)
		}
	}
	return err
}

func (p *Phase) addMonitor(monitor *Monitor) {
	p.GetRunner().getRoot().addMonitor(monitor)

}

func (p *Phase) run() error {

	idPrefix := p.GetRunner().GetId()

	// If the phase has no runner, let's create one, without a *testing.T.  This
	// allows the runner to be used disconneced from the testing module.  This
	// way, Actions can be composed using a Phase
	runner := p.GetRunner()
	if runner == nil {
		p.Runner = &Run{}
		p.DefaultRunDealer.Runner = p.Runner
		p.savedRunner = p.Runner
		runner = p.GetRunner()
	} else {
		p.connected = true
	}

	// The Runner is public; we do not want people messing with it in the middle
	// of a Run
	if p.previousRun && p.GetRunner() != p.savedRunner {
		p.Log.Printf("saved: %v, new: %v", p.savedRunner, p.GetRunner())
		return fmt.Errorf("Phase.Run configuration error: the Runner was changed")
	} else {
		p.savedRunner = p.GetRunner()
	}
	t := runner.T
	//  The testing.T on the Runner is public.  We don't want people messing with
	//  it either.
	if p.previousRun {
		if t != runner.savedT {
			return fmt.Errorf("Phase.Run configuration error: the *testing.T inside the Runner was changed")
		}
	} else {
		p.savedRunner.savedT = t
		p.previousRun = true
	}

	if t != nil {
		t.Cleanup(p.teardown)
	}

	//	if t != nil && p.Name == "" {
	//		t.Fatal("test name must be defined")
	//	}

	//	if t != nil && p.BaseRunner == nil {
	//		p.BaseRunner = &base.ClusterTestRunnerBase{}
	//	}

	if len(p.Setup) > 0 {
		for _, step := range p.Setup {
			if step.Modify != nil {
				if downerStep, ok := step.Modify.(TearDowner); ok {
					tdFunction := downerStep.Teardown()

					if tdFunction != nil {
						p.Log.Printf("[R] %v Installed auto-teardown for %T", idPrefix, step.Modify)
						p.teardowns = append(p.teardowns, downerStep.Teardown())
					}
				}
			}
			if err := processStep(t, step, &p.Log, p, SetupRunner); err != nil {
				if t != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return err
			}
			if monitorStep, ok := step.Modify.(Monitor); ok {
				p.addMonitor(&monitorStep)
				monitorStep.Monitor(p.savedRunner.ChildWithT(t, MonitorRunner))
			}
		}
	}

	var savedErr error
	if len(p.MainSteps) > 0 {
		// Is this the first phase with MainSteps for this runner?
		if !p.GetRunner().postSetup {
			for _, disruptor := range runner.getDisruptors() {
				if d, ok := disruptor.(PostMainSetupHook); ok &&
					runner.getRoot() == runner.parent && // We're just above the root
					!runner.getRoot().postMainSetupDone { // And we're first here

					runner.getRoot().postMainSetupDone = true

					log.Printf("[R] Running post-main-setup hook")
					err := d.PostMainSetupHook(runner.ChildWithT(t, HookRunner))
					if err != nil {
						runner.T.Fatalf("post-setup hook failed: %v", err)
					}
				}
			}
		}
		p.GetRunner().postSetup = true
		// log.Printf("Starting main steps")
		for _, step := range p.MainSteps {
			if err := processStep(t, step, &p.Log, p, StepRunner); err != nil {
				savedErr = err
				if t != nil && step.Name != "" {
					// TODO: Interact:
					// - continue (ignore error)
					// - hold (show time left for the test)
					// - kill (run no teardown)
					// - finish (run teardowns; go to next test if available)
					t.Errorf("[R] %v test failed: %v", idPrefix, err)
				}
				// TODO this should be pluggable
				//p.BaseRunner.DumpTestInfo(p.Name)
				break
			}
			if monitorStep, ok := step.Modify.(Monitor); ok {
				p.addMonitor(&monitorStep)
				monitorStep.Monitor(p.savedRunner)
			}
		}
	}

	if t == nil {
		// If we're not running under testing.T's supervision, we need to run
		// the teardown ourselves.

		p.teardown()
	}
	return savedErr
}

// TODO: thought for later.  Could a user control the order of individual teardowns (automatic
// and explicit) by using different phases?
func (p *Phase) teardown() {
	t := p.GetRunner().T
	// TODO: if both p.Teardown and p.teardowns were the same interface, this could be
	// a single loop.  Or: leave the individual tear downs to t.Cleanup

	if len(p.Teardown) > 0 {
		p.Log.Printf("Starting teardown")
		// This one runs in normal order, since the user listed them themselves
		for i, step := range p.Teardown {
			if err := processStep(t, step, &p.Log, p, TearDownRunner); err != nil {
				if t == nil {
					p.Log.Printf("Tear down step %d failed: %v", i, err)
				} else {
					t.Errorf("teardown failed: %v", err)
				}
				// We do not return here; we keep going doing whatever
				// teardown we can
			}
		}
	}

	if len(p.teardowns) > 0 {
		// TODO move this to t.Cleanup and make it depend on t != nil?
		// This one runs in reverse order, since they were added by the setup steps
		p.Log.Printf("Starting auto-teardown for %s", t.Name())
		for i := len(p.teardowns) - 1; i >= 0; i-- {
			td := p.teardowns[i]
			p.Log.Printf("[R] Teardown: %T", td)
			if err := td.Execute(); err != nil {
				if t == nil {
					p.Log.Printf("auto-teardown failed: %v", err)
				} else {
					t.Errorf("auto-teardown failed: %v", err)
				}
				// We do not return here; we keep going doing whatever
				// teardown we can
			}
		}
	}
}

type RunDealer interface {
	SetRunner(parent *Run, kind RunnerType)
	GetRunner() *Run
}

type DefaultRunDealer struct {
	// Perhaps make this private?
	// For now it is public to break less things
	Runner *Run
}

func (d *DefaultRunDealer) SetRunner(parent *Run, kind RunnerType) {
	if parent == nil {
		d.Runner = nil
		return
	}
	r := parent.ChildWithT(parent.T, kind)
	d.Runner = r
}

func (d *DefaultRunDealer) GetRunner() *Run {
	return d.Runner
}
