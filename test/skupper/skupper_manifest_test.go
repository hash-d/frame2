package skupper

import (
	"fmt"
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/validate"
)

func TestSkupperManifest(t *testing.T) {
	r := &frame2.Run{
		T: t,
	}

	expected := []validate.SkupperManifestContentImage{
		{
			Name:       "quay.io/skupper/skupper-router:main",
			Repository: "https://github.com/skupperproject/skupper-router",
		},
		{
			Name:       "quay.io/skupper/service-controller:master",
			Repository: "https://github.com/skupperproject/skupper",
		},
		{
			Name:       "quay.io/skupper/config-sync:master",
			Repository: "https://github.com/skupperproject/skupper",
		},
		{
			Name:       "quay.io/skupper/flow-collector:master",
			Repository: "https://github.com/skupperproject/skupper",
		},
		{
			Name:       "quay.io/prometheus/prometheus:v2.42.0",
			Repository: "",
		},
	}

	for _, e := range expected {
		individualPhase := frame2.Phase{
			Runner: r,
			Doc:    fmt.Sprintf("Checks that %q is being checked individually, and also for error", e.Repository),
			MainSteps: []frame2.Step{
				{
					Doc: "Positive check",
					Validator: &validate.SkupperManifest{
						Path: "testdata/manifest.json",
						Expected: validate.SkupperManifestContent{
							Images: []validate.SkupperManifestContentImage{
								{
									Name:       e.Name,
									Repository: e.Repository,
								},
							},
						},
					},
				}, {
					// Today, this is overkill, as we do not check Repository.  In practice, it checks that
					// :noexpected many times, with no additional checks
					Doc: "Negative check",
					Validator: &validate.SkupperManifest{
						Path: "testdata/manifest.json",
						Expected: validate.SkupperManifestContent{
							Images: []validate.SkupperManifestContentImage{
								{
									Name:       "quay.io/skupper/skupper-router:notexpected",
									Repository: e.Repository,
								},
							},
						},
					},
					ExpectError: true,
				},
			},
		}

		individualPhase.Run()
	}
}
