package frame2

import (
	"fmt"
	"regexp"
	"strings"
)

// A way to verify cli commands output
//
// StdOut and StdErr take a slice of plain strings.  It will expect that
// each string comes after the previous one.  In other words, the search
// for the second string from StdOut starts where the match for the first
// one finished.
//
// If you want to search for one static string instead, where there is
// nothing in between each segment, just use a single item with one big
// string
//
// StdOutRe and StdErrRe take a slice of regular expressions.  Those do
// not have the same restriction on one coming after the other.  If you
// want that behavior with regexes, create a single regex with the two
// expressions you're looking for.
//
// StdOutReNot and StdErrReNot behave like the previous ones, but ensure
// that the patters are not there in the checked string
type Expect struct {
	StdOut      []string
	StdErr      []string
	StdOutRe    []regexp.Regexp
	StdErrRe    []regexp.Regexp
	StdOutReNot []regexp.Regexp
	StdErrReNot []regexp.Regexp
}

// Looks for each bit (a substring), inside the string s, in order
//
// The 'name' is used for the error message.
func checkPlain(s string, bits []string, name string) (err error) {
	var startPos int
	missingPieces := []string{}

	for _, item := range bits {
		partial := s[startPos:]

		index := strings.Index(partial, item)
		if index >= 0 {
			// We found something, so the next check will start
			// where that match finished
			startPos += index + len(item)
		} else {
			missingPieces = append(missingPieces, item)
			// we continue even if an error, to report all missing pieces
		}
	}

	if len(missingPieces) > 0 {
		if len(bits) == 1 {
			err = fmt.Errorf(
				"Expected %v: \n%s\n",
				name,
				bits[0],
			)
		} else {
			msg := fmt.Sprintf(
				"Expected %v:\n%s\nmissing bits:\n",
				name,
				strings.Join(bits, "(...)"),
			)
			for _, mp := range missingPieces {
				msg = fmt.Sprintf("%v- %v\n", msg, mp)
			}
			err = fmt.Errorf(msg)
		}
	}
	return
}

// Looks for each bit (a regular expression), inside the string s.  Each bit
// is checked against the whole string, so they can be in a different order
// in the string.
//
// If expected is true; a bit that does not match will be an error; if it is
// false, a bit that matches will be an error.
//
// The 'name' is used for the error message.
func checkRe(s string, bits []regexp.Regexp, name string, expected bool) (err error) {

	var problems []string

	for _, b := range bits {
		match := b.MatchString(s)
		if match && !expected {
			problems = append(problems, fmt.Sprintf("Unexpected %s: regular expression %v matched", name, &b))
		}

		if !match && expected {
			problems = append(problems, fmt.Sprintf("Expected %s not found: regular expression %v did not match", name, &b))
		}
	}

	if len(problems) > 0 {
		message := fmt.Sprintf("Errors checking regular expressions on %v:\n", name)
		for _, p := range problems {
			message += fmt.Sprintf("- %s\n", p)
		}
		err = fmt.Errorf(message)
	}

	return err
}

// Groups and reports on a set of errors for the same input
func groupErrors(name, actual string, errors []error) (err error) {

	var hasErrors bool
	for _, e := range errors {
		if e != nil {
			hasErrors = true
			break
		}
	}
	if !hasErrors {
		return
	}
	message := "Incorrect output:\n"
	for _, e := range errors {
		if e != nil {
			message += fmt.Sprintf("%v\n", e)
		}
	}
	message += fmt.Sprintf("Actual %v:\n%v", name, actual)

	return fmt.Errorf(message)
}

// Checks all items from the specification.
func (e Expect) Check(stdout, stderr string) (err error) {

	stdOutErrors := groupErrors(
		"stdout",
		stdout,
		[]error{
			checkPlain(stdout, e.StdOut, "stdout"),
			checkRe(stdout, e.StdOutRe, "stdout", true),
			checkRe(stdout, e.StdOutReNot, "stdout", false),
		})
	stdErrErrors := groupErrors(
		"stderr",
		stderr,
		[]error{
			checkPlain(stderr, e.StdErr, "stderr"),
			checkRe(stderr, e.StdErrRe, "stderr", true),
			checkRe(stderr, e.StdErrReNot, "stderr", false),
		})

	var message string
	if stdOutErrors != nil {
		message += fmt.Sprint(stdOutErrors)
	}
	if stdErrErrors != nil {
		message += fmt.Sprint(stdErrErrors)
	}

	if message != "" {
		err = fmt.Errorf(message)
	}
	return

}
