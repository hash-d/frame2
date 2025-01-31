package f2ocp

// Try to keep this file in sync with k8s_deployment

import (
	"context"
	"fmt"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"time"

	osappsv1 "github.com/openshift/api/apps/v1"
	clientset "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// See DeploymentConfigCreate for a more complete interface
type DeploymentConfigCreateSimple struct {
	Name           string
	Namespace      *f2k8s.Namespace
	DeploymentOpts f2k8s.DeploymentOpts
	Wait           time.Duration // Waits for the deployment to be ready.  Otherwise, returns as soon as the create instruction has been issued.  If the wait lapses, return an error.

	Ctx context.Context

	Result *osappsv1.DeploymentConfig

	frame2.DefaultRunDealer
}

//          Image         string
//          Labels        map[string]string
//          RestartPolicy v12.RestartPolicy
//          Command       []string
//          Args          []string
//          EnvVars       []v12.EnvVar
//          ResourceReq   v12.ResourceRequirements
//          SecretMounts  []SecretMount

// TODO: remove this whole thing?
func (d *DeploymentConfigCreateSimple) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)

	var volumeMounts []v1.VolumeMount
	var volumes []v1.Volume

	for _, v := range d.DeploymentOpts.SecretMounts {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      v.Name,
			MountPath: v.MountPath,
		})
		volumes = append(volumes, v1.Volume{
			Name: v.Name,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: v.Secret,
				},
			},
		})
	}

	// Container to use
	containers := []v1.Container{
		{
			Name:            d.Name,
			Image:           d.DeploymentOpts.Image,
			ImagePullPolicy: v1.PullAlways,
			Env:             d.DeploymentOpts.EnvVars,
			Resources:       d.DeploymentOpts.ResourceReq,
			VolumeMounts:    volumeMounts,
		},
	}

	// Customize commands and arguments if any informed
	if len(d.DeploymentOpts.Command) > 0 {
		containers[0].Command = d.DeploymentOpts.Command
	}
	if len(d.DeploymentOpts.Args) > 0 {
		containers[0].Args = d.DeploymentOpts.Args
	}

	deploymentconfig := &osappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.Name,
		},
		Spec: osappsv1.DeploymentConfigSpec{

			Selector: d.DeploymentOpts.Labels,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      d.Name,
					Namespace: d.Namespace.GetNamespaceName(),
					Labels:    d.DeploymentOpts.Labels,
				},
				Spec: v1.PodSpec{
					Volumes:       volumes,
					Containers:    containers,
					RestartPolicy: d.DeploymentOpts.RestartPolicy,
				},
			},
			Replicas: 1,
		},
	}

	phase := frame2.Phase{
		Runner: d.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &DeploymentConfigCreate{
					Namespace:        d.Namespace,
					DeploymentConfig: deploymentconfig,
					Ctx:              ctx,
				},
			},
		},
	}
	err := phase.Run()
	if err != nil {
		return fmt.Errorf("failed to create deploymentconfig: %w", err)
	}

	if d.Wait > 0 {
		ctx := d.Ctx
		var fn context.CancelFunc
		if ctx == nil {
			ctx, fn = context.WithTimeout(context.Background(), d.Wait)
			defer fn()
		}

		phase := frame2.Phase{
			Runner: d.Runner,
			MainSteps: []frame2.Step{
				{
					Validator: &DeploymentConfigValidate{
						Namespace:        d.Namespace,
						Name:             d.Name,
						MinReadyReplicas: 1,
					},
					ValidatorRetry: frame2.RetryOptions{
						Ctx:        ctx,
						KeepTrying: true,
					},
				},
			},
		}
		return phase.Run()
	}

	return nil
}

// Executes a fully specified OCP DeploymentConfig
//
// # See DeploymentConfigCreateSimple for a simpler interface
type DeploymentConfigCreate struct {
	Namespace        *f2k8s.Namespace
	DeploymentConfig *osappsv1.DeploymentConfig

	Result *osappsv1.DeploymentConfig
	Ctx    context.Context
}

func (d *DeploymentConfigCreate) Execute() error {
	ctx := frame2.ContextOrDefault(d.Ctx)

	var err error
	client, err := clientset.NewForConfig(d.Namespace.GetKubeConfig().GetRestConfig())
	if err != nil {
		return fmt.Errorf("Failed to obtain clientset")
	}
	d.Result, err = client.DeploymentConfigs(d.Namespace.GetNamespaceName()).Create(
		ctx,
		d.DeploymentConfig,
		metav1.CreateOptions{},
	)

	if err != nil {
		return fmt.Errorf("Failed to create deploymentconfig: %w", err)
	}

	return nil
}

type DeploymentConfigValidate struct {
	Namespace        *f2k8s.Namespace
	Name             string
	MinReadyReplicas int
	Ctx              context.Context

	Result *osappsv1.DeploymentConfig

	frame2.Log
	frame2.DefaultRunDealer
}

func (d *DeploymentConfigValidate) Validate() error {
	ctx := frame2.ContextOrDefault(d.Ctx)

	client, err := clientset.NewForConfig(d.Namespace.GetKubeConfig().GetRestConfig())

	d.Result, err = client.DeploymentConfigs(d.Namespace.GetNamespaceName()).Get(
		ctx,
		d.Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("Failed to get deploymentconfig %q: %w", d.Name, err)
	}

	if int(d.Result.Status.ReadyReplicas) < d.MinReadyReplicas {
		return fmt.Errorf(
			"DeploymentConfig %q has only %d ready replicas (expected %d)",
			d.Name,
			d.Result.Status.ReadyReplicas,
			d.MinReadyReplicas,
		)
	}

	return nil
}

// Wait for the named deploymentconfig to be available.  By default, it
// waits for up to two minutes, and ensures that the deployment reports
// as ready for at least 10s.
//
// That behavior can be changed using the RetryOptions field. On that
// field, the Ctx field cannot be set; if a different timeout is desired,
// set it on the Action's Ctx itself, and it will be used for the
// RetryOptions.
type DeploymentConfigWait struct {
	Name      string
	Namespace *f2k8s.Namespace
	Ctx       context.Context

	// On this field, do not set the context.  Use the DeploymentConfigWait.Ctx,
	// instead, it will be used for the underlying Retry
	RetryOptions frame2.RetryOptions
	frame2.DefaultRunDealer
	*frame2.Log
}

func (w DeploymentConfigWait) Validate() error {
	if w.RetryOptions.Ctx != nil {
		panic("RetryOptions.Ctx cannot be set for DeploymentConfigWait")
	}
	retry := w.RetryOptions
	if retry.IsEmpty() {
		ctx, cancel := context.WithTimeout(frame2.ContextOrDefault(w.Ctx), time.Minute*2)
		defer cancel()
		retry = frame2.RetryOptions{
			Ctx:        ctx,
			KeepTrying: true,
			Ensure:     10,
		}
	}
	phase := frame2.Phase{
		Runner: w.GetRunner(),
		Doc:    fmt.Sprintf("Waiting for deploymentconfig %q on ns %q", w.Name, w.Namespace.GetNamespaceName()),
		MainSteps: []frame2.Step{
			{
				// TODO: stuff within functions need their runners replaced?
				ValidatorRetry: retry,
				Validator: &f2general.Function{
					Fn: func() error {
						validator := &DeploymentConfigValidate{
							Namespace:        w.Namespace,
							Name:             w.Name,
							MinReadyReplicas: 1,
						}
						inner1 := frame2.Phase{
							Runner: w.GetRunner(),
							Doc:    fmt.Sprintf("Get the deploymentconfig %q on ns %q", w.Name, w.Namespace.GetNamespaceName()),
							MainSteps: []frame2.Step{
								{
									Validator: validator,
								},
							},
						}
						err := inner1.Run()
						if err != nil {
							return err
						}

						inner2 := frame2.Phase{
							Runner: w.GetRunner(),
							Doc:    fmt.Sprintf("Check that the deploymentconfig %q is ready", w.Name),
							MainSteps: []frame2.Step{
								{
									Validator: &f2general.Function{
										Fn: func() error {
											if validator.Result == nil {
												return fmt.Errorf("deploymentconfig not ready: result is nil")
											}
											if validator.Result.Status.ReadyReplicas == 0 {
												return fmt.Errorf("deploymentconfig not ready: ready replicas is 0")
											}
											return nil
										},
									},
								},
							},
						}
						return inner2.Run()
					},
				},
			},
		},
	}

	return phase.Run()
}

/*
 * TODO: currently, this is a copy of DeploymentCreate stuff

type OCPDeploymentConfigAnnotate struct {
	Namespace   *base.ClusterContext
	Name        string
	Annotations map[string]string

	Ctx context.Context
}

func (kda OCPDeploymentConfigAnnotate) Execute() error {
	ctx := frame2.ContextOrDefault(kda.Ctx)
	// Retrieving Deployment
	deploy, err := kda.Namespace.VanClient.KubeClient.AppsV1().Deployments(kda.Namespace.VanClient.Namespace).Get(ctx, kda.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if deploy.Annotations == nil {
		deploy.Annotations = map[string]string{}
	}

	for k, v := range kda.Annotations {
		deploy.Annotations[k] = v
	}
	_, err = kda.Namespace.VanClient.KubeClient.AppsV1().Deployments(kda.Namespace.Namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	return err

}
*/

type DeploymentConfigUndeploy struct {
	Name      string
	Namespace *f2k8s.Namespace
	Wait      time.Duration // Waits for the deployment to be gone.  Otherwise, returns as soon as the delete instruction has been issued.  If the wait lapses, return an error.

	Ctx context.Context
	frame2.DefaultRunDealer
}

func (k *DeploymentConfigUndeploy) Execute() error {
	ctx := frame2.ContextOrDefault(k.Ctx)

	client, err := clientset.NewForConfig(k.Namespace.GetKubeConfig().GetRestConfig())

	err = client.DeploymentConfigs(k.Namespace.GetNamespaceName()).Delete(
		ctx,
		k.Name,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return err
	}
	if k.Wait == 0 {
		return nil
	}
	phase := frame2.Phase{
		Runner: k.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Doc: "Confirm the deploymentconfig is gone",
				Validator: &DeploymentConfigValidate{
					Namespace: k.Namespace,
					Name:      k.Name,
					Ctx:       ctx,
				},
				ExpectError: true,
				ValidatorRetry: frame2.RetryOptions{
					Ctx:        ctx,
					Timeout:    k.Wait,
					KeepTrying: true,
				},
			},
		},
	}
	return phase.Run()
}
