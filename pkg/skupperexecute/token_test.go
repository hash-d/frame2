package skupperexecute_test

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/disruptors"
	"github.com/hash-d/frame2/pkg/environment"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/skupperexecute"
	"gotest.tools/assert"
)

func TestTokenCreate(t *testing.T) {

	runner := &frame2.Run{
		T: t,
	}
	runner.AllowDisruptors([]frame2.Disruptor{
		&disruptors.UpgradeAndFinalize{},
		&disruptors.KeepWalking{},
	})

	tokenFile := fmt.Sprintf("/tmp/toktok-%d", rand.Intn(1000))

	phase := frame2.Phase{
		Runner: runner,
		MainSteps: []frame2.Step{
			{
				Name: "basic",
				Modify: &TokenCheck{
					FileName: tokenFile,
				},
			}, {
				Name: "defaults",
				Modify: &TokenCheck{
					FileName:      tokenFile,
					CheckDefaults: true,
				},
			}, {
				Name: "expiry",
				Modify: &TokenCheck{
					FileName: tokenFile,
					Expiry:   "60m",
				},
			}, {
				Name: "expiry-with-max",
				Modify: &TokenCheck{
					FileName:       tokenFile,
					Expiry:         "60m",
					MaxExpiryDelta: time.Minute * 10,
				},
			}, {
				Name: "expiry-with-impossible-max",
				Modify: &TokenCheck{
					FileName:       tokenFile,
					Expiry:         "60m",
					MaxExpiryDelta: time.Nanosecond * 1,
					ExpectError:    true,
				},
			}, {
				Name: "name",
				Modify: &TokenCheck{
					FileName: tokenFile,
					Name:     "asdf",
				},
			}, {
				Name: "password",
				Modify: &TokenCheck{
					FileName: tokenFile,
					Password: "asdf",
				},
			}, {
				Name: "tokenType",
				Modify: &TokenCheck{
					FileName:      tokenFile,
					TokenType:     "claim",
					CheckDefaults: true,
				},
			}, {
				Name: "uses",
				Modify: &TokenCheck{
					FileName: tokenFile,
					Uses:     "2",
				},
			},
		},
	}
	assert.Assert(t, phase.Run())
}

type TokenCheck struct {
	FileName string

	// TODO consider replacing the repetition of fields
	// by an embedded field
	Expiry    string
	Name      string
	Password  string
	TokenType string
	Uses      string

	MaxExpiryDelta time.Duration
	ExpectError    bool

	CheckDefaults bool

	frame2.DefaultRunDealer
	frame2.Log
}

func (t *TokenCheck) Teardown() frame2.Executor {
	return execute.Function{
		Fn: func() error {
			log.Printf("TearDown: Removing token file %q", t.FileName)
			return os.Remove(t.FileName)
		},
	}
}

func (tc TokenCheck) Execute() error {
	if tc.FileName == "" {
		panic("I need a file to write and read")
	}
	envSetup := &environment.JustSkupperSingle{
		Name:         "token-create-test",
		AutoTearDown: true,
	}

	// We could have a single namespace and create the all the tokens on it.
	// However, we'd not be able to test disruptors such as the upgrade with it
	setup := frame2.Phase{
		Runner: tc.Runner,
		Doc:    "Create a skupper installation",
		Setup: []frame2.Step{
			{
				Modify: envSetup},
		},
	}
	if err := setup.Run(); err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	// We do not care whether pub or private; just pick the first one
	// (there should be only one)
	ns := envSetup.Topo.ListAll()[0]

	basicWait := frame2.RetryOptions{
		Timeout:    time.Minute * 2,
		KeepTrying: true,
	}

	phase := frame2.Phase{
		Runner: tc.Runner,
		Setup: []frame2.Step{
			{
				Modify: &skupperexecute.TokenCreate{
					Namespace: ns,
					Expiry:    tc.Expiry,
					Name:      tc.Name,
					Password:  tc.Password,
					TokenType: tc.TokenType,
					Uses:      tc.Uses,
					FileName:  tc.FileName,
				},
			},
		},
		MainSteps: []frame2.Step{
			{
				Validator: &skupperexecute.TokenCheck{
					Namespace:      ns,
					Expiry:         tc.Expiry,
					Name:           tc.Name,
					Password:       tc.Password,
					TokenType:      tc.TokenType,
					Uses:           tc.Uses,
					FileName:       tc.FileName,
					CheckDefaults:  tc.CheckDefaults,
					MaxExpiryDelta: tc.MaxExpiryDelta,
				},
				ValidatorRetry:    basicWait,
				ValidatorSubFinal: true,
				ExpectError:       tc.ExpectError,
			},
		},
	}
	return phase.Run()

}
