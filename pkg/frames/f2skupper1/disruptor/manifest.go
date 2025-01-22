package disruptor

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"
)

type SkipManifestCheck struct {
}

func (s SkipManifestCheck) DisruptorEnvValue() string {
	return "SKIP_MANIFEST_CHECK"
}

func (s *SkipManifestCheck) Inspect(step *frame2.Step, phase *frame2.Phase) {
	for _, v := range step.GetValidators() {
		if v, ok := v.(*f2skupper1.SkupperManifest); ok {
			v.SkipComparison = true
		}
	}
}
