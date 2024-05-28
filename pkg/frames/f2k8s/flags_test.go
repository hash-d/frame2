package f2k8s

import (
	"flag"
	"fmt"
	"os"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"gotest.tools/assert"
)

func TestFlags(t *testing.T) {
	fmt.Println(contexts)
	assert.Assert(t, ConnectInitial())
}

func TestMain(m *testing.M) {
	frame2.Flag()
	Flag()
	flag.Parse()
	os.Exit(m.Run())
}
