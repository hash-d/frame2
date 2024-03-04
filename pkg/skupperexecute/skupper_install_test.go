package skupperexecute_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/disruptors"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
	"github.com/hash-d/frame2/pkg/skupperexecute"
	"github.com/hash-d/frame2/pkg/subrunner"
	"github.com/hash-d/frame2/pkg/topology"
	"github.com/hash-d/frame2/pkg/topology/topologies"
	"github.com/hash-d/frame2/pkg/validate"
	"github.com/skupperproject/skupper/test/utils/base"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

func TestSkupperInstallEffects(t *testing.T) {

	baseRunner := &base.ClusterTestRunnerBase{}
	runner := &frame2.Run{
		T: t,
	}
	runner.AllowDisruptors([]frame2.Disruptor{
		&disruptors.UpgradeAndFinalize{},
		&disruptors.ConsoleAuth{},
		&disruptors.SkipManifestCheck{},
		&disruptors.KeepWalking{},
	})

	var topo topology.Basic

	topo = &topologies.Single{
		Name:              "skupper-install-effects",
		TestRunnerBase:    baseRunner,
		Type:              topology.Public,
		SkipSkupperDeploy: true,
	}
	build := topology.TopologyBuild{
		Topology:     &topo,
		AutoTearDown: true,
	}

	setup := frame2.Phase{
		Runner: runner,
		Doc:    "Create a single, skupper-less namespace",
		Setup: []frame2.Step{
			{
				Modify: &build,
			},
		},
	}
	assert.Assert(t, setup.Run())

	ns, err := topo.Get(topology.Public, 1)
	assert.Assert(t, err)

	basicWait := frame2.RetryOptions{
		Timeout:    time.Minute * 2,
		KeepTrying: true,
	}

	phase := subrunner.Effects[skupperexecute.CliSkupperInstall, *skupperexecute.CliSkupperInstall]{
		ExecutionProfile: subrunner.BOTH,
		BaseFrame: &skupperexecute.CliSkupperInstall{
			Namespace: ns,
		},
		TearDown: []frame2.Step{
			{
				Modify: &skupperexecute.SkupperDelete{
					Namespace: ns,
				},
				Validator: &k8svalidate.Pods{
					Namespace: ns,
					Labels: map[string]string{
						"app.kubernetes.io/part-of": "skupper",
					},
					ExpectNone: true,
				},
				ValidatorRetry: frame2.RetryOptions{
					KeepTrying: true,
					Timeout:    time.Minute * 2,
				},
			},
		},
		Combos: map[string][]string{
			// Ensures all deployable resources are present, and check annotations
			"annotations-full": []string{"console", "annotations", "console-auth-openshift"},
		},
		// TODO: move all validators below to individual frames.
		// - Something like validateskupper.Console{}
		// Those frames should
		// - Be version-aware
		// - Allow to pick what type of testing to run (filter)
		//   - K8S resource verification
		//   - Application effect (eg curl)
		//   - or: white box vs black box
		Effects: map[string]subrunner.CauseEffect[skupperexecute.CliSkupperInstall]{
			"defaults": {
				// Do not do combos with this one, as its validations might conflict
				// with other CauseEffect items
				Doc: "Confirm a plain skupper install is successful and with expected default values",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       false,
					EnableFlowCollector: false,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.Pods{
						Namespace: ns,
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "skupper",
						},
						ExpectExactly:         2,
						NegativeContainerList: []string{"flow-collector"},
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values: map[string]string{
							"console":                "false",
							"console-authentication": "internal",
							"flow-collector":         "false",
						},
					},
				},
			},
			"annotations": {
				Doc: "Annotations must be present on all elements created by Skupper",
				Patch: skupperexecute.CliSkupperInstall{
					Annotations: []string{
						"skupper.io/qe-test=true",
						"skupper.io/qe-test-name=annotations",
					},
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.Pods{
						Namespace: ns,
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "skupper",
						},
						OtherAnnotations: map[string]string{
							"skupper.io/qe-test":      "true",
							"skupper.io/qe-test-name": "annotations",
						},
					},
					// TODO: others, too?  Services, etc
				},
			},
			"flow-collector": {
				Doc: "Check flow collector without the console",
				Patch: skupperexecute.CliSkupperInstall{
					EnableFlowCollector: true,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.Pods{
						Namespace:     ns,
						Labels:        map[string]string{"skupper.io/component": "service-controller"},
						ContainerList: []string{"flow-collector"},
						ExpectExactly: 1,
					},
					&k8svalidate.Pods{
						Namespace: ns,
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "skupper",
						},
						ExpectMin: 3,
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values: map[string]string{
							"console":        "false",
							"flow-collector": "true",
						},
					},
				},
			},
			"console": {
				Doc: "Check a basic console",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.Pods{
						Namespace:     ns,
						Labels:        map[string]string{"skupper.io/component": "service-controller"},
						ExpectPhase:   corev1.PodRunning,
						ContainerList: []string{"flow-collector"},
						ExpectExactly: 1,
					},
					&k8svalidate.Pods{
						Namespace: ns,
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "skupper",
						},
						ExpectMin: 3,
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values: map[string]string{
							"console":        "true",
							"flow-collector": "true",
						},
					},
				},
			},
			"console-user": {
				Doc: "Confirms that the console user can be configured",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
					ConsoleUser:         "testuser",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.SecretGet{
						Namespace: ns,
						Name:      "skupper-console-users",
						Keys:      []string{"testuser"},
					},
				},
			},
			"console-password": {
				Doc: "Confirms that the console password can be configured",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
					ConsolePassword:     "testpassword",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.SecretGet{
						Namespace: ns,
						Name:      "skupper-console-users",
						Expect:    map[string][]byte{"admin": []byte("testpassword")},
					},
				},
			},
			"console-user-password": {
				Doc: "Confirms that the console user and password can be configured",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
					ConsoleUser:         "testuser",
					ConsolePassword:     "testpassword",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.SecretGet{
						Namespace: ns,
						Name:      "skupper-console-users",
						Expect:    map[string][]byte{"testuser": []byte("testpassword")},
					},
				},
			},
			"console-auth-internal": {
				Doc: "Confirms that the console authentication can be set to internal",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
					ConsoleAuth:         "internal",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.SecretGet{
						Namespace: ns,
						Name:      "skupper-console-users",
						Keys:      []string{"admin"},
					},
				},
			},
			"console-auth-unsecured": {
				Doc: "Confirms that the console authentication can be set to unsecured",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
					ConsoleAuth:         "unsecured",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.SecretGet{
						Namespace:    ns,
						Name:         "skupper-console-users",
						ExpectAbsent: true,
					},
				},
			},
			"console-auth-openshift": {
				Doc: "Confirms that the console authentication can be set to openshift",
				Patch: skupperexecute.CliSkupperInstall{
					EnableConsole:       true,
					EnableFlowCollector: true,
					ConsoleAuth:         "openshift",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.SecretGet{
						Namespace:    ns,
						Name:         "skupper-console-users",
						ExpectAbsent: true,
					},
					&validate.Container{
						Namespace:     ns,
						PodSelector:   validate.ServiceControllerSelector,
						ContainerName: "oauth-proxy",
					},
				},
			},
			"network-policy": {
				Doc: "Skupper init is able to create its NetworkPolicy",
				Patch: skupperexecute.CliSkupperInstall{
					CreateNetworkPolicy: true,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.NetworkPolicy{
						Namespace: ns,
						Name:      "skupper",
					},
				},
			},
			"cluster-permissions": {
				// TODO: this may give a false negative, if the cluster
				// already had the ClusterRoleBinding before.  Currently,
				// they do not get removed on skupper delete.
				//
				// See
				// - https://github.com/skupperproject/skupper/issues/813
				// - https://github.com/skupperproject/skupper/issues/857
				Doc: "Skupper init can enable cluster-wide permissions.  Attention to false negatives",
				Patch: skupperexecute.CliSkupperInstall{
					EnableClusterPermissions: true,
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.ClusterRoleBindingGet{
						Namespace: ns,
						Name: fmt.Sprintf(
							"skupper-service-controller-extended-%s",
							ns.Namespace,
						),
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values:    map[string]string{"cluster-permissions": "true"},
					},
				},
			},
			"site-name": {
				Doc: "Skupper init set a site name",
				Patch: skupperexecute.CliSkupperInstall{
					SiteName: "custom-site-name",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values:    map[string]string{"name": "custom-site-name"},
					},
				},
			},
			"router-logging": {
				Doc: "Skupper init set a site name",
				Patch: skupperexecute.CliSkupperInstall{
					RouterLogging: "trace",
				},
				ValidatorsRetry: basicWait,
				Validators: []frame2.Validator{
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-site",
						Values:    map[string]string{"router-logging": "trace"},
					},
					&k8svalidate.ConfigMap{
						Namespace: ns,
						Name:      "skupper-internal",
						CMValidator: func(cm v1.ConfigMap) error {
							// TODO: move this to a dedicated frame?
							if config, ok := cm.Data["skrouterd.json"]; ok {
								var root []any
								err := json.Unmarshal([]byte(config), &root)
								if err != nil {
									return fmt.Errorf("failed to get root list: %w", err)
								}
								for _, item := range root {
									if item, ok := item.([]any); ok {
										if len(item) != 2 {
											return fmt.Errorf("unexpected number of items on the structure: %d", len(item))
										}
										if name, ok := item[0].(string); ok {
											if name != "log" {
												continue
											}
											if details, ok := item[1].(map[string]interface{}); ok {
												if level, ok := details["enable"]; ok {
													if level, ok := level.(string); ok {
														if level == "trace+" {
															return nil
														} else {
															return fmt.Errorf("router logging is configured for unexpected value %q", level)
														}
													} else {
														return fmt.Errorf("value is not a string")
													}
												} else {
													return errors.New("unable to find 'enable' key on map")
												}
											} else {
												return errors.New("second item on the structure is not a map")
											}
										} else {
											return errors.New("first item on the structure is not a string")
										}
									} else {
										return errors.New("List is not composed of sublists")
									}
								}
							}
							return errors.New("'log' configuration not found")
						},
					},
				},
			},
			//
		},
	}.GetPhase(runner)

	assert.Assert(t, phase.Run())

}
