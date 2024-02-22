package k8svalidate

import (
	"context"
	"fmt"
	"strings"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Pods struct {
	Namespace *base.ClusterContext
	Labels    map[string]string
	Ctx       context.Context

	// Ignored if zero
	ExpectMin int

	// Ignored if zero
	ExpectMax int

	// Ignored if zero
	ExpectExactly int

	// Expect no results
	ExpectNone bool

	//ExpectCondition corev1.PodConditionType
	ExpectPhase corev1.PodPhase

	// Function to run against each pod, to validate.  Provides a way
	// to execute more complex validations not available above
	PodValidator func(corev1.Pod) error

	// A complex validation on the list as a whole.  Allows, for example
	// to aggregate values from the different pods for verification
	ListValidator func([]corev1.Pod) error

	// List of labels expected to be set on the Pod, regardless of value
	LabelList []string

	// List of labels expected to _not_ be set on the Pod, regardless of value
	NegativeLabelList []string

	// Other labels expected to be on the Pod, besides those used on the
	// selector, with their expected values
	OtherLabels map[string]string

	// Labels listed on this map must not have the mapped value.  If the
	// label does not exist at all on the pod, the NegativeLabels test is
	// considered successful, unless NegativeLabelsExist is set to true
	NegativeLabels map[string]string

	NegativeLabelsExist bool

	// List of annotations expected to be set on the Pod, regardless of value
	AnnotationList []string

	// List of annotations expected to _not_ be set on the Pod, regardless of value
	NegativeAnnotationList []string

	// Other annotations expected to be on the Pod, besides those used on the
	// selector, with their expected values
	OtherAnnotations map[string]string

	// Annotations listed on this map must not have the mapped value.  If the
	// label does not exist at all on the pod, the NegativeAnnotations test is
	// considered successful, unless NegativeAnnotationsExist is set to true
	NegativeAnnotations map[string]string

	NegativeAnnotationsExist bool

	Result *[]corev1.Pod

	frame2.Log
	frame2.DefaultRunDealer
}

func (p *Pods) Validate() error {
	ctx := frame2.ContextOrDefault(p.Ctx)

	var items []string
	// TODO: is there an API that already does that?
	for k, v := range p.Labels {
		items = append(items, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(items, ",")
	podList, err := p.Namespace.VanClient.KubeClient.CoreV1().Pods(p.Namespace.Namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: selector,
		})
	if err != nil {
		return fmt.Errorf("failed to get pod list by labels: %w", err)
	}
	p.Result = &podList.Items

	asserter := frame2.Asserter{}

	numMatches := len(*p.Result)
	if p.ExpectNone {
		asserter.Check(numMatches == 0, "found %d pods instead of zero", numMatches)
	} else {
		if p.ExpectMin != 0 {
			asserter.Check(numMatches >= p.ExpectMin, "expected at least %d pods, found %d", p.ExpectMin, numMatches)
		}
		if p.ExpectMax != 0 {
			asserter.Check(numMatches >= p.ExpectMax, "expected at most %d pods, found %d", p.ExpectMax, numMatches)
		}
		if p.ExpectExactly != 0 {
			asserter.Check(numMatches >= p.ExpectExactly, "expected exactly %d pods, found %d", p.ExpectExactly, numMatches)
		}
	}

	if numMatches == 0 {
		return asserter.Error()
	}

	for _, pod := range *p.Result {
		if p.PodValidator != nil {
			asserter.CheckError(p.PodValidator(pod))
		}
		if p.ExpectPhase != "" {
			asserter.Check(p.ExpectPhase == pod.Status.Phase, "expected pod %q to be in phase %q, found %q", pod.Name, p.ExpectPhase, pod.Status.Phase)
		}
		// Labels
		for _, l := range p.LabelList {
			_, ok := pod.Labels[l]
			asserter.Check(ok, "label %q not found on pod %q", l, pod.Name)
		}
		for _, l := range p.NegativeLabelList {
			_, ok := pod.Labels[l]
			asserter.Check(!ok, "label %q found on pod %q, unexpectedly", l, pod.Name)
		}
		for k, v := range p.OtherLabels {
			podValue, ok := pod.Labels[k]
			if asserter.Check(ok, "label %q not found on pod %q", k, pod.Name) == nil {
				asserter.Check(
					v == podValue,
					"pod %q has value %q for label %q, while expected was %q",
					pod.Name, podValue, k, v,
				)
			}
		}
		for k, v := range p.NegativeLabels {
			podValue, ok := pod.Labels[k]
			if asserter.Check(ok || !p.NegativeLabelsExist, "label %q not found on pod %q", k, pod.Name) == nil {
				asserter.Check(
					v != podValue,
					"pod %q has unexpected value %q for label %q",
					pod.Name, podValue, k,
				)
			}
		}
		// Annotations
		for _, l := range p.AnnotationList {
			_, ok := pod.Annotations[l]
			asserter.Check(ok, "annotation %q not found on pod %q", l, pod.Name)
		}
		for _, l := range p.NegativeAnnotationList {
			_, ok := pod.Annotations[l]
			asserter.Check(!ok, "annotation %q found on pod %q, unexpectedly", l, pod.Name)
		}
		for k, v := range p.OtherAnnotations {
			podValue, ok := pod.Annotations[k]
			if asserter.Check(ok, "annotation %q not found on pod %q", k, pod.Name) == nil {
				asserter.Check(
					v == podValue,
					"pod %q has value %q for annotation %q, while expected was %q",
					pod.Name, podValue, k, v,
				)
			}
		}
		for k, v := range p.NegativeAnnotations {
			podValue, ok := pod.Annotations[k]
			if asserter.Check(ok || !p.NegativeAnnotationsExist, "annotation %q not found on pod %q", k, pod.Name) == nil {
				asserter.Check(
					v != podValue,
					"pod %q has unexpected value %q for annotation %q",
					pod.Name, podValue, k,
				)
			}
		}
	}

	if p.ListValidator != nil {
		asserter.CheckError(p.ListValidator(*p.Result))
	}

	return asserter.Error()
}
