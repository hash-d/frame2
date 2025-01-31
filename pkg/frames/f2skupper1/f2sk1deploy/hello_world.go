package f2sk1deploy

import (
	"context"
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/topology"
	v1 "k8s.io/api/core/v1"
)

// Deploys HelloWorld; frontend on pub1, backend on prv1
type HelloWorld struct {
	Topology *topology.Basic

	// This will create K8S services
	CreateServices bool

	// This will create Skupper services; if CreateServices is also
	// true, the Skupper service will be based on the K8S service.
	// Otherwise, it exposes the deployment.
	//
	// The Skupper service will use the HTTP protocol
	SkupperExpose bool

	frame2.DefaultRunDealer
}

// Deploys the hello-world-frontend pod on pub1 and hello-world-backend pod on
// prv1, and validate they are available
func (hw HelloWorld) Execute() error {

	pub, err := (*hw.Topology).Get(f2k8s.Public, 1)
	if err != nil {
		return fmt.Errorf("failed to get public-1")
	}
	prv, err := (*hw.Topology).Get(f2k8s.Private, 1)
	if err != nil {
		return fmt.Errorf("failed to get private-1")
	}

	phase := frame2.Phase{
		Runner: hw.Runner,
		Doc:    "Install Hello World front and back ends",
		MainSteps: []frame2.Step{
			{
				Doc: "Install Hello World frontend",
				Modify: &HelloWorldFrontend{
					Target:         pub,
					CreateServices: hw.CreateServices,
					SkupperExpose:  hw.SkupperExpose,
				},
			}, {
				Doc: "Install Hello World backend",
				Modify: &HelloWorldBackend{
					Target:         prv,
					CreateServices: hw.CreateServices,
					SkupperExpose:  hw.SkupperExpose,
				},
			},
		},
	}
	return phase.Run()
}

type HelloWorldBackend struct {
	Target         *f2k8s.Namespace
	CreateServices bool
	SkupperExpose  bool
	Protocol       string // This will default to http if not specified

	Ctx context.Context
	frame2.DefaultRunDealer
}

func (h *HelloWorldBackend) Execute() error {

	ctx := frame2.ContextOrDefault(h.Ctx)

	proto := h.Protocol
	if proto == "" {
		proto = "http"
	}

	labels := map[string]string{"app": "backend"}

	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Installing hello-world-backend",
				Modify: &f2k8s.DeploymentCreateSimple{
					Name:      "backend",
					Namespace: h.Target,
					DeploymentOpts: f2k8s.DeploymentOpts{
						Image:         "quay.io/skupper/hello-world-backend",
						Labels:        labels,
						RestartPolicy: v1.RestartPolicyAlways,
					},
					Ctx: ctx,
				},
			}, {
				Doc: "Creating a local service for hello-world-backend",
				Modify: &f2k8s.ServiceCreate{
					Namespace: h.Target,
					Name:      "backend",
					Labels:    labels,
					Ports:     []int32{8080},
				},
				SkipWhen: !h.CreateServices,
			}, {
				Doc: "Exposing the local service via Skupper",
				Modify: &f2skupper1.SkupperExpose{
					Namespace: h.Target,
					Type:      "service",
					Name:      "backend",
					Protocol:  proto,
				},
				SkipWhen: !h.CreateServices || !h.SkupperExpose,
			}, {
				Doc: "Exposing the deployment via Skupper",
				Modify: &f2skupper1.SkupperExpose{
					Namespace: h.Target,
					Ports:     []int{8080},
					Type:      "deployment",
					Name:      "backend",
					Protocol:  proto,
				},
				SkipWhen: h.CreateServices || !h.SkupperExpose,
				Validator: &f2k8s.DeploymentWait{
					Namespace: h.Target,
					Name:      "backend",
				},
			},
		},
	}
	return phase.Run()
}

type HelloWorldFrontend struct {
	Target         *f2k8s.Namespace
	CreateServices bool
	SkupperExpose  bool
	Protocol       string // This will default to http if not specified

	Ctx context.Context

	frame2.DefaultRunDealer
}

func (h *HelloWorldFrontend) Execute() error {

	ctx := frame2.ContextOrDefault(h.Ctx)

	proto := h.Protocol
	if proto == "" {
		proto = "http"
	}

	labels := map[string]string{"app": "frontend"}

	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Doc: "Installing hello-world-frontend",
				Modify: &f2k8s.DeploymentCreateSimple{
					Name:      "frontend",
					Namespace: h.Target,
					DeploymentOpts: f2k8s.DeploymentOpts{
						Image:         "quay.io/skupper/hello-world-frontend",
						Labels:        labels,
						RestartPolicy: v1.RestartPolicyAlways,
					},
					Ctx: ctx,
				},
			}, {
				Doc: "Creating a local service for frontend",
				Modify: &f2k8s.ServiceCreate{
					Namespace: h.Target,
					Name:      "frontend",
					Labels:    labels,
					Ports:     []int32{8080},
				},
				SkipWhen: !h.CreateServices,
			}, {
				Doc: "Exposing the local service via Skupper",
				Modify: &f2skupper1.SkupperExpose{
					Namespace: h.Target,
					Type:      "service",
					Name:      "frontend",
					Protocol:  proto,
				},
				SkipWhen: !h.CreateServices || !h.SkupperExpose,
			}, {
				Doc: "Exposing the deployment via Skupper",
				Modify: &f2skupper1.SkupperExpose{
					Namespace: h.Target,
					Ports:     []int{8080},
					Type:      "deployment",
					Name:      "frontend",
					Protocol:  proto,
				},
				Validator: &f2k8s.DeploymentWait{
					Namespace: h.Target,
					Name:      "frontend",
				},
				SkipWhen: h.CreateServices || !h.SkupperExpose,
			},
		},
	}
	return phase.Run()
}

// Validates a Hello World deployment by Curl from the given Namespace.
//
// The individual validaators (front and back) may be configured, but generally do not need to;
// they'll use the default values.
type HelloWorldValidate struct {
	Namespace               *f2k8s.Namespace
	HelloWorldValidateFront HelloWorldValidateFront
	HelloWorldValidateBack  HelloWorldValidateBack

	frame2.Log
	frame2.DefaultRunDealer
}

func (h HelloWorldValidate) Validate() error {
	if h.Namespace == nil {
		panic("HelloWorldValidate configuration error: empty Namespace")
	}

	if h.HelloWorldValidateFront.Namespace == nil {
		h.HelloWorldValidateFront.Namespace = h.Namespace
	}
	if h.HelloWorldValidateFront.Runner == nil {
		h.HelloWorldValidateFront.Runner = h.Runner
	}
	h.HelloWorldValidateFront.OrSetLogger(h.Log.GetLogger())

	if h.HelloWorldValidateBack.Namespace == nil {
		h.HelloWorldValidateBack.Namespace = h.Namespace
	}
	if h.HelloWorldValidateBack.Runner == nil {
		h.HelloWorldValidateBack.Runner = h.Runner
	}
	h.HelloWorldValidateBack.OrSetLogger(h.Log.GetLogger())

	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Validators: []frame2.Validator{
					&h.HelloWorldValidateFront,
					&h.HelloWorldValidateBack,
				},
			},
		},
	}
	phase.OrSetLogger(h.Logger)
	return phase.Run()
}

type HelloWorldValidateFront struct {
	Namespace       *f2k8s.Namespace
	ServiceName     string // default is frontend
	ServicePort     int    // default is 8080
	ServiceInsecure bool   // Ignores certificate problems
	ServiceProto    string // default is http

	frame2.Log
	frame2.DefaultRunDealer
}

func (h HelloWorldValidateFront) Validate() error {
	if h.Namespace == nil {
		return fmt.Errorf("HelloWorldValidateFront configuration error: empty Namespace")
	}
	svc := h.ServiceName
	if svc == "" {
		svc = "frontend"
	}
	port := h.ServicePort
	if port == 0 {
		port = 8080
	}
	proto := h.ServiceProto
	if proto == "" {
		proto = "http"
	}
	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Validator: &f2k8s.Curl{
					Namespace:   h.Namespace,
					Url:         fmt.Sprintf("%s://%s:%d", proto, svc, port),
					Fail400Plus: true,
					Log:         h.Log,
					CurlOptions: f2k8s.CurlOpts{
						Insecure: h.ServiceInsecure,
					},
				},
			},
		},
	}
	phase.SetLogger(h.Logger)
	return phase.Run()
}

type HelloWorldValidateBack struct {
	Namespace       *f2k8s.Namespace
	ServiceName     string // default is backend
	ServicePort     int    // default is 8080
	ServicePath     string // default is api/hello
	ServiceProto    string // default http
	ServiceInsecure bool   // ignores cert problems

	frame2.Log
	frame2.DefaultRunDealer
}

func (h HelloWorldValidateBack) Validate() error {
	if h.Namespace == nil {
		return fmt.Errorf("HelloWorldValidateBack configuration error: empty Namespace")
	}
	svc := h.ServiceName
	if svc == "" {
		svc = "backend"
	}
	port := h.ServicePort
	if port == 0 {
		port = 8080
	}
	path := h.ServicePath
	if path == "" {
		path = "api/hello"
	}
	proto := h.ServiceProto
	if proto == "" {
		proto = "http"
	}
	phase := frame2.Phase{
		Runner: h.Runner,
		MainSteps: []frame2.Step{
			{
				Validator: &f2k8s.Curl{
					Namespace:   h.Namespace,
					Url:         fmt.Sprintf("%s://%s:%d/%s", proto, svc, port, path),
					Fail400Plus: true,
					Log:         h.Log,
					CurlOptions: f2k8s.CurlOpts{
						Insecure: h.ServiceInsecure,
					},
				},
			},
		},
	}
	phase.SetLogger(h.Logger)
	return phase.Run()
}
