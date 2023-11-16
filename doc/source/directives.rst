
==========
Directives
==========

frame2 has two main parts:

-  A library of validators and executors, that promote code reuse and consistency
-  A runner that:

   -  Promotes test reuse
   -  Provides standard test 'services'

.. admonition:: Oh, but I hate tests on a struct

   That's fine:

   -  Use the struct for setup/teardown
   -  Put your test's code within an Executor
   -  Use the components as stand-alone items, if you want
   -  Have your test have a single step: calling your executor, where the test can
      be written in any way

   This way, *most* of the benefits of the framework will still be available.

   (but not all: for example, it will be difficult to reuse the testing with
   disruptors).

Directives:

-  Quick testing should be quick; simple testing should be simple; more complex testing should be straightforward

-  High-level bits ease development, but they're built on top of lower-level bits, which are available and allow for fine-grained control

-  Test code should focus on the test at hand; let the framework handle everything else

   -  Test code written this way should be easier to reuse (eg on different topologies)
   -  It should be more readable, as well

-  When semantics change on a code unit, just create a V2 (do not break other tests or spend a lot of time
   ensuring older tests are good)

-  Tests should be readable:

   -  The code should be smallish and to-the-point
   -  The log should clearly identify the steps.  A ``grep`` on such logs should be enough to extract
      a description of the test.

.

- overengineer a bit.  See execute.Cmd.  To just execute a command, they already have Go's exec.Cmd, so we add return status list checks and output checks
- Make low-level bits available.  Compose into easy to use things.
- Provide things in different levels of complexity. xSimple requires almost no configuration; xFull lets you configure everything
- As things are composed, do not duplicate fields; add the bit's struct into the bigger one; adjust as needed.  See CliSkupper
- Document and test
- Make V2 of things, if logic (better word for this?) changes are required
- Reuse as much as possible.  frame2 testing uses frame2.TestRun
