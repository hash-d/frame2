package disruptor

import (
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
)

// Enables TLS Secret Generation
type EnableTls struct{}

func (n EnableTls) DisruptorEnvValue() string {
	return "ENABLE_TLS"
}

func (u *EnableTls) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if mod, ok := step.Modify.(*f2skupper1.SkupperExpose); ok {
		mod.GenerateTlsSecrets = true
		log.Printf("ENABLE_TLS: %v", mod.Namespace.GetNamespaceName())
	}
}
