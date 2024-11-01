package f2k8s

import (
	"bytes"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Execute helps executing commands on a given pod, using the k8s rest api
// returning stdout, stderr, err
// This function is nil safe and so stdout and stderr are always returned
//
// This is copied straight from skupper test/utils/k8s/execute.go
//
// Do not use this out of f2k8s.  Today, it is exported because it is used
// by validate/curl, which will be moved to f2k8s (?) in a refactoring
//
// TODO: perhaps use a call to kubectl exec, instead; it would make the test
// more easily reproduceable manually, and we probably could get rid of
// pkg/frames/f2k8s/scheme.go & go altogether
func Execute(kubeClient kubernetes.Interface, config *rest.Config, ns string, pod, container string, command []string) (bytes.Buffer, bytes.Buffer, error) {
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
	request.VersionedParams(&v1.PodExecOptions{
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
