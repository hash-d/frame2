package f2k8s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"strings"

	frame2 "github.com/hash-d/frame2/pkg"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO move to frames.k8svalidate and refactor to share code with
// k8svalidate.Pods
//
// This will return a single pod.  If name is provided, that will be used.
// Otherwise, the pods with the given label will be listed, and the first
// in the list will be returned
type PodGet struct {
	Namespace *Namespace
	Name      string
	Labels    map[string]string
	Ctx       context.Context

	Result *corev1.Pod

	frame2.Log
	frame2.DefaultRunDealer
}

func (g *PodGet) Execute() error {
	ctx := frame2.ContextOrDefault(g.Ctx)

	if g.Name != "" {
		var err error
		g.Result, err = g.Namespace.PodInterface().Get(ctx, g.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod %q by name: %w", g.Name, err)
		}
		return nil
	}

	var items []string
	// TODO: is there an API that already does that?
	for k, v := range g.Labels {
		items = append(items, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(items, ",")
	podList, err := g.Namespace.PodInterface().List(
		ctx,
		metav1.ListOptions{
			LabelSelector: selector,
			Limit:         1,
		})
	if err != nil {
		return fmt.Errorf("failed to get pod list by labels: %w", err)
	}
	if len(podList.Items) != 1 {
		return fmt.Errorf("failed to get pod by labels")
	}
	g.Result = &podList.Items[0]

	return nil
}

type PodExecute struct {
	Pod       *PodGet
	Container string
	Command   []string
	Ctx       context.Context
	Expect    frame2.Expect // Configures checks on Stdout and Stderr

	// TODO: use common code with execute.Command
	ForceOutput bool // Shows this command's output on log, regardless of environment config
	// ForceNoOutput bool       // No output, regardless of environment config.  Takes precedence over the above

	// These are probably not implementable; k8s.io/client-go/tools/remotecommand does not
	// return exit status
	//
	// AcceptReturn  []int      // consider these return status as a success.  Default only 0
	// FailReturn    []int      // Fail on any of these return status.  Default anything other than 0

	frame2.Log

	*f2general.CmdResult
}

func (e *PodExecute) Execute() error {
	err := e.Pod.Execute()
	if err != nil {
		return fmt.Errorf("PodExecute failed to get pod: %w", err)
	}

	e.Log.Printf("Executing on pod %q: %s", e.Pod.Result.Name, e.Command)

	stdout, stderr, err := executeOnPod(
		e.Pod.Namespace.KubeClient(),
		e.Pod.Namespace.GetKubeConfig().GetRestConfig(),
		e.Pod.Namespace.GetNamespaceName(),
		e.Pod.Result.GetName(),
		e.Container,
		e.Command)
	e.CmdResult = &f2general.CmdResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if e.ForceOutput || frame2.IsVerboseCommandOutput() {
		e.Log.Printf("STDOUT:\n%v\n", stdout.String())
		e.Log.Printf("STDERR:\n%v\n", stderr.String())
		e.Log.Printf("Error: %v\n", err)
	}
	if err != nil {
		return fmt.Errorf("PodExecute failed execution:  %w", err)
	}

	expectErr := e.Expect.Check(e.CmdResult.Stdout, e.CmdResult.Stderr)
	if expectErr != nil {
		return expectErr
	}

	return nil
}

func (e *PodExecute) Validate() error {
	return e.Execute()
}

// executeOnPod helps executing commands on a given pod, using the k8s rest api
// returning stdout, stderr, err
// This function is nil safe and so stdout and stderr are always returned
//
// This is copied straight from skupper test/utils/k8s/execute.go
//
// TODO: perhaps use a call to kubectl exec, instead; it would make the test
// more easily reproducible manually, and we probably could get rid of
// pkg/frames/f2k8s/scheme.go & go altogether
func executeOnPod(kubeClient kubernetes.Interface, config *rest.Config, ns string, pod, container string, command []string) (bytes.Buffer, bytes.Buffer, error) {
	// nil safe
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return stdout, stderr, err
	}

	// k8s request to be executed as a remote command
	request := restClient.Post().
		Resource("pods").
		Namespace(ns).
		Name(pod).
		SubResource("exec")
	request.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	// Executing
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		return stdout, stderr, err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})

	// Returning
	return stdout, stderr, err
}
