package f2k8s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	frame2 "github.com/hash-d/frame2/pkg"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// If CurlOptions.Timeout is zero, a default is set, instead.
type Curl struct {
	Namespace *Namespace

	// A default of 60s is set for the timeout, if the original value is
	// zero.
	CurlOptions CurlOpts
	Url         string
	Fail400Plus bool
	Podname     string // Passed to tools.Curl.  Generally safe to leave empty.  Check tools.Curl docs
	DeployCurl  bool

	// TODO: Add cli.Expect to inspect results?
	frame2.Log
}

func (c Curl) Validate() error {
	if c.DeployCurl {
		deployCurl(c.Namespace.KubeClient(), c.Namespace.GetNamespaceName(), "curl")
		waitPhase := frame2.Phase{
			MainSteps: []frame2.Step{
				{
					Validator: &f2general.Executor{
						Executor: &PodGet{
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
	c.Log.Printf("Calling Curl from %q on %q", c.Podname, c.Url)
	resp, err := curl(
		c.Namespace.KubeClient(),
		c.Namespace.GetKubeConfig().GetRestConfig(),
		c.Namespace.GetNamespaceName(),
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

// CurlOpts allows specifying arguments to run curl on a pod
type CurlOpts struct {
	Silent   bool
	Insecure bool
	Username string
	Password string
	Timeout  int
	Verbose  bool
}

// ToParams returns curl options serialized as a string slice
func (c *CurlOpts) ToParams() []string {
	params := []string{}
	if c.Silent {
		params = append(params, "-s")
	}
	if c.Verbose {
		params = append(params, "-v")
	}
	if c.Insecure {
		params = append(params, "-k")
	}
	if c.Username != "" && c.Password != "" {
		params = append(params, "-u", fmt.Sprintf("%s:%s", c.Username, c.Password))
	}
	if c.Timeout > 0 {
		params = append(params, "--max-time", strconv.Itoa(c.Timeout))
		params = append(params, "--connect-timeout", strconv.Itoa(c.Timeout))
	}
	return params
}

// CurlResponse wraps a response for a curl execution
type CurlResponse struct {
	HttpVersion  string
	StatusCode   int
	ReasonPhrase string
	Headers      map[string]string
	Body         string
	Output       string
}

// Curl runs curl on a given pod (or if empty, it will try to find
// the service-controller pod and run against it).
func curl(kubeClient kubernetes.Interface, config *restclient.Config, ns, podName, url string, opts CurlOpts) (*CurlResponse, error) {
	var pod *v1.Pod
	var err error

	// Initializing the response
	response := &CurlResponse{
		Headers: map[string]string{},
	}

	if podName == "" {
		// If podName not provided try running against the skupper-controller podName
		podList, err := kubeClient.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "skupper.io/component=service-controller",
			Limit:         1,
		})
		if err != nil {
			log.Println("error retrieving pod list")
			return response, err
		}
		if len(podList.Items) != 1 {
			log.Println("no pods to run curl")
			return response, fmt.Errorf("no pods to run curl")
		}
		pod = &podList.Items[0]
	} else {
		// Retrieving the given pod
		pod, err = kubeClient.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			log.Printf("unable to find pod: %s", podName)
			return response, err
		}
	}

	suffix, _ := uuid.NewUUID()
	headersFile := fmt.Sprintf("/tmp/curl.Headers.%s", suffix.String())
	bodyFile := fmt.Sprintf("/tmp/curl.Body.%s", suffix.String())

	// Preparing command to run
	command := []string{"curl", "-D", headersFile, "-o", bodyFile}
	command = append(command, opts.ToParams()...)
	command = append(command, url)

	// Executing through the API
	curlDoneCh := make(chan struct{})
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10
	}
	timeoutCh := time.After(time.Duration(timeout) * time.Second)
	var stderr bytes.Buffer

	go func() {
		_, stderr, err = executeOnPod(kubeClient, config, ns, pod.Name, pod.Spec.Containers[0].Name, command)
		close(curlDoneCh)
	}()

	// wait on curl to finish or a timeout to happen
	select {
	case <-curlDoneCh:
		break
	case <-timeoutCh:
		return response, fmt.Errorf("timed out waiting on curl")
	}

	// Curl's Output (not the Body, but regular Output or errors)
	response.Output = stderr.String()

	if err != nil {
		log.Printf("error executing curl: %s", err)
		return response, err
	}

	// Reading response Body
	stdout, stderr, err := executeOnPod(kubeClient, config, ns, pod.Name, pod.Spec.Containers[0].Name, []string{"cat", bodyFile})
	if err != nil {
		log.Printf("error retrieving response Body - %s", stderr.String())
		return nil, err
	}
	response.Body = stdout.String()

	// Reading header file
	stdout, stderr, err = executeOnPod(kubeClient, config, ns, pod.Name, pod.Spec.Containers[0].Name, []string{"cat", headersFile})
	if err != nil {
		log.Printf("error retrieving Output Headers - %s", stderr.String())
		return nil, err
	}

	// Parsing Headers
	statusLine := true
	for {
		line, err := stdout.ReadString('\n')
		if err == io.EOF {
			break
		}

		// Parsing the HTTP status line
		if statusLine {
			statusLine = false
			httpStatusLine := strings.Split(line, " ")
			if len(httpStatusLine) < 3 {
				return nil, fmt.Errorf("error parsing HTTP status line - not enough elements [%d] - expected "+
					"(at least) 3 - line: %s", len(httpStatusLine), httpStatusLine)
			}
			response.HttpVersion = httpStatusLine[0]
			response.StatusCode, err = strconv.Atoi(httpStatusLine[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing HTTP status code '%s' - error: %s", httpStatusLine[1], err)
			}
			response.ReasonPhrase = strings.Join(httpStatusLine[2:], " ")
			continue
		}

		// Processing Headers
		header := strings.Split(line, ":")
		if line == "" || len(header) != 2 {
			continue
		}
		response.Headers[header[0]] = header[1]
	}

	// Removing the Output files
	_, stderr, err = executeOnPod(kubeClient, config, ns, pod.Name, pod.Spec.Containers[0].Name, []string{"rm", headersFile, bodyFile})
	if err != nil {
		log.Printf("error removing Headers and Body files - %s", stderr.String())
		return nil, err
	}

	// All seems good
	return response, nil
}

// deployCurl helps deploying a Pod that provides "curl"
// you must wait for it to be ready
func deployCurl(kubeClient kubernetes.Interface, ns, pod string) (*v1.Pod, error) {
	terminationPeriodSecs := int64(30)
	return kubeClient.CoreV1().Pods(ns).Create(context.TODO(), &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   pod,
			Labels: map[string]string{"app": "curl"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "curl", Image: "curlimages/curl", Command: strings.Split("tail -f /dev/null", " ")},
			},
			RestartPolicy:                 v1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: &terminationPeriodSecs,
		},
	}, metav1.CreateOptions{})
}
