package disruptor

import (
	"github.com/hash-d/frame2/pkg/frames/f2skupper1"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
)

// This disruptor will cause any services created with http or http2 as
// their protocol to use tcp instead.
type NoHttp struct {
}

func (n NoHttp) DisruptorEnvValue() string {
	return "NO_HTTP"
}

func (n *NoHttp) Inspect(step *frame2.Step, phase *frame2.Phase) {
	// TODO change this by some kind of interface, so different service
	//      create types can be used (different UI)?

	// log.Printf("[D] NO_HTTP inspecting %T", step.Modify)
	if mod, ok := step.Modify.(*f2skupper1.SkupperServiceCreate); ok {
		if mod.Protocol == "http" || mod.Protocol == "http2" {
			log.Printf("[D] NO_HTTP overriding service %q as 'tcp'", mod.Name)
			mod.Protocol = "tcp"
		}
	}
	if mod, ok := step.Modify.(*f2skupper1.SkupperExpose); ok {
		log.Printf("Considering SkupperExpose %q", mod.Name)
		if mod.Protocol == "http" || mod.Protocol == "http2" {
			log.Printf("[D] NO_HTTP overriding service %q as 'tcp'", mod.Name)
			mod.Protocol = "tcp"
		}
	}
}
