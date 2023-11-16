
============
Requirements
============

Random
======

-  A new testing image, with generic code for running validations (such as the HelloWorld tester, etc).  Also, load it with the stuff that helps on testing (curl, netcat, openssl and so on)

-  Allow for a 'preferred' topology.  That is, by default, the test does not want to run on Simplest, but it is also ok to run on Simplest (or any other)

Validators & Executors/Modifiers
================================

These are reusable pieces of code that can be composed into more complex functionality
or actual tests.  They follow an Action pattern and can be used on the Runner (where
they get special 'integrated' functionality), or as stand-alone components.

.. sidebar:: Associated validators

   An example of validators associated with executors is for Skupper init.  When running
   `skupper init`, it is normal to wait for the pods to be up before proceeding to the
   next step.

   We shouldn't need to specify that for every test.  Instead, the component that does
   the skupper init should have such validations already out of the box.

*Validators* simply check for some status, while *Executors* are expected to change
some state.  Executors may have associated validators as well.

.. todo:: should there be an option to disable an associated validator?

   Or should they be written in such a way that a user wanting to skip the validation
   could pick another component that does just the init part (and which is used by
   the more generic one, in turn)?


-  At the lowest level, execute system commands or K8S or other APIs to perform a task

-  One level above, these can be grouped or abstracted:

   -  For example, a "SkupperOp" abstraction can use different lower-level bits to perform
      a task, using different interfaces (CLI, API/VanClient, Ansible, site-controller), depending
      on run-time configuration

   -  Or a group of ordered tasks to link two skupper installations (ie, one to generate
      the token, the other to create the link)

-  At the highest level, time savers:

   -  Create namespaces and linked Skupper installations in a given topology
   -  Install an application (echo server, Hipster) on that topology
   -  Run a full, standardized test against such an environment (for example, Hipster's
      load generator)

.. note:: there was a step/walk distinction before...

      Where walk would be a grouping of executors/validators.

      However, that needs re-worked.  In special, avoiding dependency cycles.

      Perhaps, create a 'leave' package, with all the lowest-level components.
      Or leave validator/executor for that, and create a new package that
      consumes them.

There are several types of executors:

:Unit: they only do what they're supposed to do: no validations, no composition

:Ensure: execute a unit, and run only enough validation to ensure that it's completed.
         For example: skupper init + wait for pods to be up (and perhaps run skupper
         version or something).  Or check that the dns name for a service is present
         on its own namespace (and possibly on other connected namespaces as well)

:SkupperOp: define a Skupper operation, and allow it to be executed in different ways,
          according to the test configuration (cli, annotations, API, Ansible, site-controler...)

          Writing tests using these allow them to be promptly re-used to test different
          interfaces

:Tester: a minimal test, similar to InitTester & co from cli testing.  They are configured
         to perform a change (Executor) and right away a set of checks (Validators).

         They can be mixed and matched just like on the cli testing.

:Topology: Create a number of namespaces, install Skupper into them and link them
           according to an expected topology

:Deploy: Install an application: Hello World, Echo server, Hypster.  It receives a reference
         to the topology and works with it to decide where to deploy its stuff (ie: where's
         the primary frontend/backend context?  which other public/private context are available?)

         The code should also make available some application-related tests.  For example,
         HelloWorldValidateFrontend, that would just send a Curl to it.

:Environment: Topology + Deploy.  Hello World on hub/spoke, for example

              Simpler tests should use it, and it should have simple and configurable alternatives:

              - environment.HelloWorldSimple: prv1 connects to pub1, front on pub, back on prv1.  Option to create services
              - e.HelloWorld. Options.  Configure deployment with a deploy.HelloWorld


Within these types, there may be two subtypes:

:Simple: A little configuration as possible.  Uses 'Full', below
:Full: complete configuration available

.. todo:: add examples

Validators and Executors may also implement the AutoDoc interface.  In that case,
the step documentation string is generated automatically from the configurations
that were provided.  Otherwise, a Doc string must be provided (or the test will fail)


Tester
======

.. todo:: check whether this really makes sense as a separate thing, or is just a
          type of executor with validators

Individual tests composed of PreValidate, Modify, PostValidate.  They can also be configured (such as a service name)

Some times, reusing bits of tests may be interesting


Auto actions
------------

The executors can install automatic features:

-  Auto teardown, to remove what's just been installed at the end of the test
-  Monitors, that compare memory usage or pod restarts between the start and end of the test, or that check
   for K8S or Skupper events
-  Debug helpers/dumpers, that react to test failures with debug information.  Skupper init installs
   a ``skupper debug dump`` helper; namespace creation installs a K8S namespace debug helper with
   events, pod status, etc.  Errors here are reported but ignored.
-  Continuous tests

Continuous tests
""""""""""""""""

These run in parallel to the actual tests.  They are informed by the runner as steps are
executed, and then report at the end with percentages.

For example, during the execution of step X, Y% of the continous tests failed (ie, between the
step starting and all its validators reporting ok).

This would allow for checking application availability while operations are done on Skupper.

For example: during a step that reproduced a DC migration, how many requests failed?

Runner
======

-  It should be simple and straightforward to skip the setup and/or teardown of tests
   properly written using the runner

-  The runner also provides some operation modes that simplify the debugging of failing tests:

   -  Interrupt on error (wait for user interaction, leaving the environment as found when
      the error ocurred

   -  Interactive: like on a debugger: step in, over or out of test steps.


Steps
-----

On a runner, each test step may have:

-  A documentation string that is reported on the test's output
-  A pre-validation step, to ensure the environment is good for running the actual test [#pre-validate]_
-  A Modify step that changes the state of the system at test
-  One or more Validator steps, that ensure the Modify step achieved its goal

   -  Validator steps may be marked as 'final'.  In that case, the same validation is run again
      at the end of the test, and the same result is expected.
   -  They can also be configured for running in different situations (``ALWAYS``, ``LONG``, etc)

-  One or more substeps, that allow for some structuring of the execution
-  Configuration:

   -  Retry options for the validations
   -  Retry options for the substeps
   -  Error expectation


.. [#pre-validate] This is useful especially for bigger executors, that expect a certain
                   environment for their execution.

                   Pre-validation steps may also be skipped for performance.

Run and Phases
--------------

A Run (name may change) is the highest-level Go ``testing.T`` test in the framework.  It is
composed of one more Phases.

On the logs, each Run produces a unique identifier string, so that test can be identified
even if the logs were moved out of their original places.

Each phase has:

-  A name and documentation string
-  Setup and Teardown steps
-  Main steps, where the actual test runs

The main reason for the Run to be split in Phases is due to the table-driven nature of
the framework.  If everything was in the same table, later items would not be able to
reference objects created in earlier items (as the whole table is evaluated at the same
time, some references would not yet be valid).  Having the phases removes that problem.

Additionally, phases can be put within loops or other control structures, allowing for
more complex behavior.

Skipping
-----------

Actions can return frame2.SkipContinue.  It means to finish the current test, but mark it as skipped in the end.

This is for the case of  using different methods to accomplish a function.

For example, we could have a SkupperOp set of actions, that call Cli, annotation or site-controller actions depending on an environment variable, but fall back on a default method if they do not have other set up.  In that case, the test using that SkupperOp would use SkipContinue

This way, the step would:

- Run, as further tests may depend on them, but on fallback
- Not count as Successes.  Skipped items could be inspected for missing alternatives on SkupperOp, in this case.


frame2.Skip.  An error.  Indicates the test should stop at that point and be skipped.  Use, for example, when requesting a different disruptor or different topology test, but the test does not support it.

A Sentinel error.  Allows for tests to indicate to the Runner that they should be skipped.

Disruptors
==========

Disruptors allow for a test to be reused in situations that are not their 'Standard temperature
and pressure' conditions.

There are several types of disruptors

-  In-between steps
-  Continuous

Upgrade
-------

When running a test with an Upgrade disruptor in the most fine-grained detail:

-  Set the test up
-  Run all steps until first Skupper-specific test
-  Validate it
-  Run Skupper upgrade
-  Re-run last validation, with increased repeat configuration
-  Continue test
-  Teardown
-  Restart, but now until the *second* Skupper-specific test
-  Repeat until all Skupper-specific tests are completed

Here, 'Skupper-specific' is a task such as Skupper init, service create, link, etc.

This would allow a very detailed Skupper upgrade test (not to be run daily, as it will
take a lot of time), reusing tests that were created for other reasons.

Pod Killer
----------

For each step of the test:

-  Run any pre-validations
-  Run the modify step
-  Run the validations
-  Force restart the pods (all or specific)
-  Re-run the validations, with additional retry configuration
-  Move to the next step

Network unavailability
----------------------

For each step:

-  Run any pre-validations
-  Make the network unavailable
-  Run the modify step
-  Wait?
-  Re-enable the network
-  Run the validations

.. todo:: check on how to make the network unavailable:

   -  NetworkPolicy

Skupper disruptor
-----------------

A continuous disruptor, where a set of skupper operations that are not
related to the test at hand and are note expected to interfere with it
are continuously executed.

For example, while Hello World is running, this disruptor could be adding
and removing services that are not part of the Hello World test suite.

The idea is to see how Skupper behaves when the control plane is constantly
busy.

Alternate version
-----------------

Skupper is initialized with one version of the skupper command, but the
ensuing access is done with a different version.

| What happens when you use RHAI 1.0 on a RHAI 1.1 installation, and vice-versa.


It should be ok if it just fails... But if it silently breaks things, we'd
need to avoid the cross-use.

Topology changes
----------------

There are two ideas here:  on the first, the topology simply keeps changing,
and we watch how Skupper reacts to that.  (continuous)

On the second, we shut down just enough skupper nodes to break the connection
between pub1 and prv1, and act like the Network disruptor, otherwise.  If the
connection is direct, skip the test.


Topologies
==========

Most tests should not define a topology.  They should only request the framework for the
minimum they require.  If all they need is pub1 and prv1, they should always be good.

For most tests, the topology will be the simplest (a segment betwen pub1 and prv1).  However,
the test can be configured to be executed in a different topology, so the tests can
be reused for topology testing.

When the test at hand is topology-specific, the test can specify it, but it won't be
used in the multiple-topology tests.

In any topology provided by the framework, one thing is constant: pub1 and prv1 are as
far away as possible from each other as the topology allows.

:Segment: public-1, private-1
:N: prv1 -> pub2 <- prv2 -> pub1.  Good for minimal multiple link testing
:Diamod: private-2 and -3 can exemplify DMZ servers, or just routing.  They do not connect to each other.  Private-1 connects to them, they connect to public-1
:Segments: consecutive segments (linear topology). public-1 at one side, private-1 at the other.  How to control direction of links?  What will other items be (public or private)
:Circle: at least three (a triangle)
:Hiperconnect: at least four; all connected
:Hub and Spokes: private-1 at the center; publics around it.
:Hubs and Spokes: Some topology at the center, with edge nodes connected to its elements
:Complex: configure with a structure
:Stich: Pubs on one DC, privs on the other.  Each node has at most two connections, both going to the other side.  It's a zig-zag with connections going always out of prv to pub

Tests do not need to tell exactly what they want.  Instead, they can ask the runner to give them something with a public-1 and a private-1 (or something else).  So, same tests can be run with different configurations.  Is this already implemented in Base?

When executing a test, the topology will be documented with an URL such as the one
below:

https://dreampuf.github.io/GraphvizOnline/#digraph%20G%20%7B%0A%0A%20prv1%20-%3E%20pub1%0A%20prv2%20-%3E%20pub2%0A%20prv1%20-%3E%20pub2%0A%20%0A%7D

Skupper info commands
=====================

While the test is running, continuousy run skupper info commands (such as ``link status`` or ``network status`` and check for failures and long-running times)

Test types
==========

Quick Scenarios
---------------

These are test templates: they have the code to setup an environment (topology + deployment), based on some
application (Kafka, HelloWorld, RabbitMQ, etc).

The idea is to have something quick to start an actual test with, or to reproduce a customer
situation.

They should be provided in "Simple" and "Full" flavors.


Longevity
---------

Tests developed for longevity testing should:

-  As much as possible, be topology-independent
-  Have very clear setup/teardown steps
-  Have the main steps running on a loop

By default, the test runs on a loop of two iterations.  This allows the
test to be checked to be good as a longevity testing.  That is, that it's
not making changes that leave the environment unsuitable for a second run.

When run as a longevity test, it will first be executed with a configuration
to do no tear down, and then repeatedly without either setup or teardown.

In these runs, the '2' loop can be maintained, or changed to just '1' (or
any other number, if that makes sense).

Questions
=========

-  Today, we have ``Execute()``, ``Validate()`` and ``Run()``.  Should they all be
   just one ``Run()`` interface, instead?  This way, Executors could be put on the
   validators part and vice-versa.  Is that good or bad?

-  ``ClusterContextPromise`` is an artifact of how frame2 evolved.  Perhaps strip that?  Or
   make a more generic thing?  Perhaps provide default values, but allow steps to
   override.

Log
===

-  Each step should be identified on the logs with its name, documentation and a
   structured number that helps identify where it stands in the overal test

   -  The numbering should go back to ``1.`` every time a new test or subtest
      starts, to keep them small

   -  This structure could also be shown at the end of the test, as test documentation
      (though simple grepping should review it)

Future
======

-  Generate manual reproducers from the tests?
-  Generate test tree/description
-  Grab Ctrl+C; if tty, give some options:

   -  Run teardown and exit
   -  Run this test's teardown and move to the next
   -  Panic
   -  Exit without tear down

   Meanwhile, the teardowns have not been run, and the user can inspect the environment at the point where Ctrl+C was hit.
