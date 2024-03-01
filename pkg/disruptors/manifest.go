package disruptors

import (
	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/skupperexecute"
)

type SkipManifestCheck struct {
}

func (s SkipManifestCheck) DisruptorEnvValue() string {
	return "SKIP_MANIFEST_CHECK"
}

func (s *SkipManifestCheck) Inspect(step *frame2.Step, phase *frame2.Phase) {
	for _, v := range step.GetValidators() {
		if v, ok := v.(*skupperexecute.SkupperManifest); ok {
			v.SkipComparison = true
		}
	}
}
