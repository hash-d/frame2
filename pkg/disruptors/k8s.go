package disruptors

import (
	"fmt"
	"log"
	"strings"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	corev1 "k8s.io/api/core/v1"
)

// Any namespaces created via execute.TestRunnerCreateNamespace will receive
// annotations that force  PSA on restricted mode
//
// TODO: In the future, make this more configurable (ie, different settings for
// different levels, or ensure no PSA at all)
//
// By default, the version is defined as "latest", but it can be configured with
// the disruptor configuration "version"
type PodSecurityAdmission struct {
	Version string // default "latest"
}

func (n PodSecurityAdmission) DisruptorEnvValue() string {
	return "PSA"
}

// For now, only a single configuration is accepted: 'version', which is
// used for all three levels.  If not set, 'latest' is used on Inspect
func (p *PodSecurityAdmission) Configure(config string) error {
	definition := strings.Split(config, "=")
	if len(definition) != 2 {
		return fmt.Errorf("%q is not a valid PSA configuration", definition)
	}
	k, v := definition[0], definition[1]
	if k == "version" {
		p.Version = v
	} else {
		return fmt.Errorf("%q is not a valid PSA configuration", k)
	}

	return nil
}

func (u *PodSecurityAdmission) Inspect(step *frame2.Step, phase *frame2.Phase) {
	version := u.Version
	if version == "" {
		version = "latest"
	}
	if mod, ok := step.Modify.(*execute.TestRunnerCreateNamespace); ok {
		if mod.Labels == nil {
			mod.Labels = make(map[string]string)
		}

		mod.Labels["security.openshift.io/scc.podSecurityLabelSync"] = "false"
		mod.Labels["pod-security.kubernetes.io/warn"] = "restricted"
		mod.Labels["pod-security.kubernetes.io/warn-version"] = version
		mod.Labels["pod-security.kubernetes.io/audit"] = "restricted"
		mod.Labels["pod-security.kubernetes.io/audit-version"] = version
		mod.Labels["pod-security.kubernetes.io/enforce"] = "restricted"
		mod.Labels["pod-security.kubernetes.io/enforce-version"] = version

		log.Printf("PSA: %v", mod.Namespace.Namespace)
	}
}

// PSADeployment will modify any deployments created by the test
// code to make them conform the the K8S PSA requirements.
type PSADeployment struct{}

func (d PSADeployment) DisruptorEnvValue() string {
	return "PSA_DEPLOYMENT"
}
func (d *PSADeployment) Inspect(step *frame2.Step, phase *frame2.Phase) {
	// log.Printf("PSA Inspecting %T", step.Modify)
	if mod, ok := step.Modify.(*execute.K8SDeployment); ok {

		_true := true

		podSpec := &mod.Deployment.Spec.Template.Spec
		if podSpec.SecurityContext == nil {
			podSpec.SecurityContext = &corev1.PodSecurityContext{}
		}
		podSpec.SecurityContext.RunAsNonRoot = &_true
		podSpec.SecurityContext.SeccompProfile = &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		}

		for i := range mod.Deployment.Spec.Template.Spec.Containers {
			container := &mod.Deployment.Spec.Template.Spec.Containers[i]

			if container.SecurityContext == nil {
				container.SecurityContext = &corev1.SecurityContext{}

			}
			container.SecurityContext.AllowPrivilegeEscalation = new(bool)

			if container.SecurityContext.Capabilities == nil {
				container.SecurityContext.Capabilities = &corev1.Capabilities{}

			}

			container.SecurityContext.Capabilities.Drop = []corev1.Capability{
				"ALL",
			}

			// We do not try to check whether something already exists and add;
			// we overwritte.  Only the Capability below should be allowed for
			// a PSA-enabled namespace
			container.SecurityContext.Capabilities.Add = []corev1.Capability{
				"NET_BIND_SERVICE",
			}
		}

		log.Printf("PSA_DEPLOYMENT: %v", mod.Namespace.Namespace)
	}
}
