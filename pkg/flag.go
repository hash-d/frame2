package frame2

import (
	"flag"
	"os"
)

// Right now, this will only add the -H flag for
// showing the flag's help (flag.Usage())
func Flag() {
	flag.BoolFunc(
		"H",
		"this help screen",
		func(string) error {
			flag.Usage()
			os.Exit(0)
			return nil
		},
	)
}
