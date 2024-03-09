// This package contains simple but full tests encapsulated
// into callable frame2 executors.
//
// They do their setup (namespace creation, skupper and app
// installation, etc) and run their validations, just like
// any other tests.
//
// They can be used in a few situations:
//
//   - When a test forces a disruptor, and all it's interested
//     is a simple test with that disruptor
//
//   - When running a test in a loop.  For example, when exploring
//     the effects of
package test_unit
