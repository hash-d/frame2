package validate

import (
	"fmt"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/tools"
)

// Provides an interface to tools.Curl, with some enhancements.
//
// If CurlOptions.Timeout is zero, a default is set, instead.
type Curl struct {
	Namespace *base.ClusterContext

	// CurlOptions is passed as-is to tools.Curl, with the exception that a
	// default of 60s is set for the timeout, if the original value is
	// zero.
	CurlOptions tools.CurlOpts
	Url         string
	Fail400Plus bool
	Podname     string // Passed to tools.Curl.  Generally safe to leave empty.  Check tools.Curl docs
	DeployCurl  bool

	// TODO: Add cli.Expect to inspect results?
	frame2.Log
}

func (c Curl) Validate() error {
	if c.DeployCurl {
		tools.DeployCurl(c.Namespace.VanClient.KubeClient, c.Namespace.Namespace, "curl")
		waitPhase := frame2.Phase{
			MainSteps: []frame2.Step{
				{
					Validator: &Executor{
						Executor: &execute.K8SPodGet{
							Namespace: c.Namespace,
							Name:      "curl",
						},
					},
					ValidatorRetry: frame2.RetryOptions{
						Timeout: 2 * time.Minute,
					},
				},
			},
		}
		err := waitPhase.Run()
		if err != nil {
			return fmt.Errorf("failed waiting for Curl pod: %w", err)
		}
	}
	if c.CurlOptions.Timeout == 0 {
		// There is no reason to give Curl no time to respond
		c.CurlOptions.Timeout = 60
	}
	c.Log.Printf("Calling Curl on %v", c.Url)
	resp, err := tools.Curl(
		c.Namespace.VanClient.KubeClient,
		c.Namespace.VanClient.RestConfig,
		c.Namespace.Namespace,
		c.Podname,
		c.Url,
		c.CurlOptions,
	)
	if resp == nil {
		c.Log.Printf("- No response from Curl")
	} else {
		c.Log.Printf("- Output:\n%v", resp.Output)
	}
	if err != nil {
		c.Log.Printf("- Err: %v", err)
		return fmt.Errorf("curl invokation failed: %w", err)
	}

	c.Log.Printf("- status code %d", resp.StatusCode)
	c.Log.Printf("- HTTP version: %v", resp.HttpVersion)
	c.Log.Printf("- Reason phrase: %v", resp.ReasonPhrase)
	c.Log.Printf("- Headers:\n%v", resp.Headers)
	c.Log.Printf("- Body:\n%v", resp.Body)

	if c.Fail400Plus && resp.StatusCode >= 400 {
		return fmt.Errorf("curl invokation returned status code %d", resp.StatusCode)
	}

	return err
}
