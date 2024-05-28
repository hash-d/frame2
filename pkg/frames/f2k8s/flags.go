package f2k8s

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// This is a list of kubeconfig files, as parsed from the
// command line flags, by ClusterType
var contexts = map[ClusterType][]string{}

// ParseContext will parse the provided value, with the
// format [kind=]/path/to/file.
//
// If kind is not provided, it will be set as 'pub' by
// default.
//
// It is used by Flag(), to add the --kube flag to the
// available flags, and that is its main objective.
//
// One can, however, programatically call it from individual
// tests to add kubeconfig files.
//
// Be aware, however, that this will affect a package
// variable (ie, it will affect any other tests in the
// same run)
func ParseContext(value string) error {
	split := strings.SplitN(value, "=", 2)
	var domain ClusterType
	var file string
	if len(split) == 1 {
		domain = "pub"
		file = split[0]
	} else {
		domain = ClusterType(split[0])
		file = split[1]
	}

	s, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to open kubeconfig file %q: %w", file, err)
	}
	if !s.Mode().IsRegular() {
		return fmt.Errorf("path %q is not a regular file", file)
	}
	contexts[domain] = append(contexts[domain], file)
	return nil
}

// Flag adds the flag "-kube" to the list of flags to be parsed
// on the command line; call it right before flag.Parse(), on
// TestMain, if you want your test to interact with multiple
// Kubernetes clusters.
func Flag() {
	flag.Func(
		"kube",
		""+
			"`path` to a kubeconfig file.\n"+
			"Use like --kube=pub=/path/tofile",
		ParseContext,
	)
}
