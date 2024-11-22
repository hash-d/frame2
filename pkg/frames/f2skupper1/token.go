package f2skupper1

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/jmespath/go-jmespath"
	"sigs.k8s.io/yaml"
)

type TokenCreate struct {
	Namespace *f2k8s.Namespace

	Expiry    string
	Name      string
	Password  string
	TokenType string
	Uses      string

	FileName string

	// By default, the token file is automatically removed at the
	// end of the test
	SkipTearDown bool

	frame2.DefaultRunDealer
	frame2.Log
	SkupperVersionerDefault
}

// TODO: replace this by f2k8s.Namespace
func (t TokenCreate) GetNamespace() string {
	return t.Namespace.GetNamespaceName()
}

func (t *TokenCreate) Teardown() frame2.Executor {
	if t.SkipTearDown {
		return nil
	}
	return f2general.Function{
		Fn: func() error {
			log.Printf("Removing file %q", t.FileName)
			return os.Remove(t.FileName)
		},
	}
}

func (tc *TokenCreate) Execute() error {

	args := []string{"token", "create"}

	if tc.Expiry != "" {
		args = append(args, "--expiry", tc.Expiry)
	}
	if tc.Name != "" {
		args = append(args, "--name", tc.Name)
	}
	if tc.Password != "" {
		args = append(args, "--password", tc.Password)
	}
	if tc.TokenType != "" {
		args = append(args, "--token-type", tc.TokenType)
	}
	if tc.Uses != "" {
		args = append(args, "--uses", tc.Uses)
	}
	if tc.FileName != "" {
		args = append(args, tc.FileName)
	}

	phase := frame2.Phase{
		Runner: tc.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					Args:        args,
					F2Namespace: tc.Namespace,
				},
			},
		},
	}
	return phase.Run()
}

// Verify that a Token (the file and its Kubernetes representation
// as a secret) matches some expectations.
type TokenCheck struct {
	Namespace *f2k8s.Namespace

	FileName string

	// If Name is not provided, it will be derived from the token
	// file.  If both provided, TokenCheck will ensure they match
	Name string

	Expiry    string
	Password  string
	TokenType string
	Uses      string

	// By default, on the comparison between the Expiry above and
	// what's on the secret, the following rules will be followed:
	//
	// - If both are below 60s, they'll be considered equivalent
	// - Otherwise, if the difference between them is less than 20%,
	// they'll be considered equivalent
	//
	// This is to allow for
	//
	// - clock differences between computer running the test and the cluster
	// - time between creating a token and calling TokenCheck
	//
	// If MaxExpiryDelta is set, however, a fixed delta will be used
	// for the comparison
	MaxExpiryDelta time.Duration

	CheckDefaults bool

	frame2.DefaultRunDealer
	frame2.Log
}

func (tc TokenCheck) Validate() error {

	expiry := tc.Expiry
	tokenType := tc.TokenType
	uses := tc.Uses

	if tc.CheckDefaults {
		if expiry == "" {
			expiry = "15m"
		}
		if tokenType == "" {
			tokenType = "claim"
		}
		if uses == "" {
			uses = "1"
		}
	}

	if tc.FileName+tc.Name == "" {
		panic("I need a file or a secret name")
	}

	var filePassword string
	var secretName string
	if tc.FileName != "" {
		// Open the token file and inspect it.
		f, err := os.Open(tc.FileName)
		if err != nil {
			return fmt.Errorf("failed to open token file: %w", err)
		}
		defer f.Close()

		contents, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read token file: %w", err)
		}

		var parsed interface{}
		err = yaml.Unmarshal(contents, &parsed)
		if err != nil {
			return fmt.Errorf("failed to parse token file: %w", err)
		}

		secretNameI, err := jmespath.Search("metadata.name", parsed)
		if err != nil {
			return fmt.Errorf("failed to apply jmes search: %w", err)
		}

		var ok bool
		if secretName, ok = secretNameI.(string); !ok {
			return fmt.Errorf("failed to get the secret name as a string (got %T, instead)", secretNameI)
		}
		log.Printf("Secret name: %q", secretName)
		if tc.Name != "" && tc.Name != secretName {
			return fmt.Errorf("The provided scret name %q is different from the one on the token %q", tc.Name, secretName)
		}

		filePasswordI, err := jmespath.Search("data.password", parsed)
		if err != nil {
			return fmt.Errorf("failed to search for password: %w", err)
		}

		if pass, ok := filePasswordI.(string); ok {
			base64pass := strings.NewReader(pass)
			decoder := base64.NewDecoder(base64.StdEncoding, base64pass)
			passwordBytes, err := io.ReadAll(decoder)
			if err != nil {
				fmt.Errorf("failed to decode file password: %w", err)
			}
			filePassword = string(passwordBytes)
		} else {
			return fmt.Errorf("failed to get the password from the file (got %T)", filePasswordI)
		}

		if tc.Password != "" {
			if filePassword != tc.Password {
				return fmt.Errorf("Given password is different than file contents")
			}

		}

		if tokenType != "" {
			resp, err := jmespath.Search(
				fmt.Sprintf(`metadata.labels."skupper.io/type" | @ == 'token-%s'`, tokenType),
				parsed,
			)
			if err != nil {
				fmt.Errorf("failed to check type on file: %w", err)
			}
			if boolResp, ok := resp.(bool); ok {
				if !boolResp {
					return fmt.Errorf("skupper.io/type label on file does not match expected value")
				}
			} else {
				return fmt.Errorf("skupper.io/type check returned non-bool response (%T)", resp)
			}
		}
	}
	validators := []frame2.Validator{}

	if secretName != "" {

		annotationValues := map[string]string{}
		labelValues := map[string]string{}
		expect := map[string][]byte{}

		if uses != "" {
			annotationValues["skupper.io/claims-remaining"] = uses
		}

		if tokenType != "" {
			labelValues["skupper.io/type"] = fmt.Sprintf("token-%s-record", tokenType)
		}

		switch {
		case tc.Password != "":
			expect["password"] = []byte(tc.Password)
		case filePassword != "":
			expect["password"] = []byte(filePassword)
		}

		var validatorFunc func(map[string]string) error
		if expiry != "" {
			validatorFunc = func(m map[string]string) error {

				strSecretExpiry := m["skupper.io/claim-expiration"]
				dateSecretExpiry, err := time.Parse(time.RFC3339, strSecretExpiry)
				if err != nil {
					return fmt.Errorf("failed parsing claim-expiration %q: %w", strSecretExpiry, err)
				}
				toExpiry := dateSecretExpiry.Sub(time.Now())
				duration, err := time.ParseDuration(expiry)
				durationS := duration.Seconds()
				toExpiryS := toExpiry.Seconds()
				deltaS := math.Abs(durationS - toExpiryS)
				if tc.MaxExpiryDelta != 0 {
					if deltaS > tc.MaxExpiryDelta.Seconds() {
						return fmt.Errorf("expiry did not match: %s (Δ=%vs)", strSecretExpiry, deltaS)
					} else {
						return nil
					}
				} else {
					if toExpiryS < 60 && durationS < 60 {
						return nil
					}
					diffRate := math.Abs(toExpiryS/durationS - 1)
					if diffRate > 20 {
						return fmt.Errorf("expiry delta greater than 20%%: %v", strSecretExpiry)
					}

				}
				return nil
			}
		}

		validators = append(validators, f2k8s.SecretGet{
			Namespace: tc.Namespace,
			Name:      secretName,
			Expect:    expect,
			Annotations: f2general.MapCheck{
				Values:       annotationValues,
				MapValidator: validatorFunc,
			},
			Labels: f2general.MapCheck{
				Values: labelValues,
			},
		})

	}

	validatePhase := frame2.Phase{
		Runner: tc.Runner,
		MainSteps: []frame2.Step{
			{
				Validators:        validators,
				ValidatorSubFinal: true,
			},
		},
	}
	return validatePhase.Run()

}
