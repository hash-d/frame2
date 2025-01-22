package f2skupper1

import (
	"context"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1/f2sk1const"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SkupperInfoContents struct {
	Images    SkupperManifestContent
	PodImages SkupperManifestContent

	HasRouter            bool
	HasServiceController bool
	HasPrometheus        bool

	RouterDeployment            *appsv1.Deployment
	ServiceControllerDeployment *appsv1.Deployment
	PrometheusDeployment        *appsv1.Deployment

	AllPods *corev1.PodList
}

// On skupper code, defined on pkg/utils/configs/manifest.go, but not as
// constants
const (
	SkupperRouterRepo = "https://github.com/skupperproject/skupper-router"
	SkupperRepo       = "https://github.com/skupperproject/skupper"
	EmptyRepo         = ""
	UnknownRepo       = "UNKNOWN"
)

// Gets various information about Skupper
// TODO: add ConfigMaps, skmanage executions
type SkupperInfo struct {
	Namespace *f2k8s.Namespace

	Result SkupperInfoContents

	Ctx context.Context
	frame2.DefaultRunDealer
	frame2.Log
}

func (s *SkupperInfo) Validate() error {
	ctx := frame2.ContextOrDefault(s.Ctx)

	var err error

	// Router deployment
	s.Result.RouterDeployment, err = s.Namespace.DeploymentInterface().Get(ctx, f2sk1const.TransportDeploymentName, metav1.GetOptions{})
	if err != nil {
		s.Log.Printf("failed to get deployment %q: %v", f2sk1const.TransportDeploymentName, err)
	} else {
		s.Result.HasRouter = true
		for _, container := range s.Result.RouterDeployment.Spec.Template.Spec.Containers {
			switch container.Name {
			case f2sk1const.TransportComponentName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: SkupperRouterRepo,
					},
				)
			case f2sk1const.ConfigSyncContainerName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,

					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: SkupperRepo,
					},
				)
			default:
				s.Log.Printf("Unknown container %q in deployment %q", container.Name, s.Result.RouterDeployment.Name)
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: UnknownRepo,
					},
				)
			}
		}

	}

	// Service Controller Deployment
	s.Result.ServiceControllerDeployment, err = s.Namespace.DeploymentInterface().Get(ctx, f2sk1const.ControllerDeploymentName, metav1.GetOptions{})
	if err != nil {
		s.Log.Printf("failed to get deployment %q: %v", f2sk1const.TransportDeploymentName, err)
	} else {
		s.Result.HasServiceController = true
		for _, container := range s.Result.ServiceControllerDeployment.Spec.Template.Spec.Containers {
			switch container.Name {
			case f2sk1const.ControllerContainerName, f2sk1const.FlowCollectorContainerName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: SkupperRepo,
					},
				)
			default:
				s.Log.Printf("Unknown container %q in deployment %q", container.Name, s.Result.RouterDeployment.Name)
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: UnknownRepo,
					},
				)
			}
		}

	}

	// Prometheus deployment
	s.Result.PrometheusDeployment, err = s.Namespace.DeploymentInterface().Get(ctx, f2sk1const.PrometheusDeploymentName, metav1.GetOptions{})
	if err != nil {
		s.Log.Printf("failed to get deployment %q: %v", f2sk1const.TransportDeploymentName, err)
	} else {
		s.Result.HasPrometheus = true
		for _, container := range s.Result.PrometheusDeployment.Spec.Template.Spec.Containers {
			switch container.Name {
			case f2sk1const.PrometheusContainerName:
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: EmptyRepo,
					},
				)
			default:
				s.Log.Printf("Unknown container %q in deployment %q", container.Name, s.Result.RouterDeployment.Name)
				s.Result.Images.Images = append(
					s.Result.Images.Images,
					SkupperManifestContentImage{
						Name:       container.Image,
						Repository: UnknownRepo,
					},
				)
			}
		}

	}

	s.Result.AllPods, err = s.Namespace.PodInterface().List(
		ctx,
		metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/part-of=skupper",
		},
	)
	if err != nil {
		s.Log.Printf("failed to get Pod list: %v", err)
	} else {
		for _, p := range s.Result.AllPods.Items {
			for _, c := range p.Spec.Containers {
				s.Result.PodImages.Images = append(
					s.Result.PodImages.Images,
					SkupperManifestContentImage{
						Name: c.Image,
						// TODO: Add respository
					},
				)
			}
		}
	}

	return nil

}
