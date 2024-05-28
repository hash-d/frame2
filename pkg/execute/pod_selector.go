package execute

import (
	"context"
	"fmt"
	"log"

	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodSelector struct {
	Namespace     *f2k8s.Namespace
	Selector      string
	ExpectNone    bool // If true, it will be an error if any pods are found
	ExpectExactly int  // if greater than 0, exactly this number of pods must be found

	// Return value
	Pods []v1.Pod
}

func (p *PodSelector) Execute() error {

	options := metav1.ListOptions{LabelSelector: p.Selector}
	podList, err := p.Namespace.PodInterface().List(context.TODO(), options)
	if err != nil {
		return err
	}
	pods := podList.Items

	log.Printf("- Found %d pod(s)", len(pods))

	if p.ExpectNone {
		if len(pods) > 0 {
			return fmt.Errorf("expected no pods, found %d", len(pods))
		}
		return nil
	} else {
		if p.ExpectExactly > 0 {
			if len(pods) != p.ExpectExactly {
				return fmt.Errorf("expected exactly %d pods, found %d", p.ExpectExactly, len(pods))
			}
		} else {
			if len(pods) == 0 {
				return fmt.Errorf("expected at least one pod, found none")
			}
		}
	}

	p.Pods = append(p.Pods, pods...)

	return nil
}
