package execute

import (
	"context"
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Executes a fully specified K8S Statefulset
type K8SStatefulSet struct {
	Namespace    *f2k8s.Namespace
	StatefulSet  *apps.StatefulSet
	AutoTeardown bool
	Ctx          context.Context

	Result *apps.StatefulSet
}

func (k *K8SStatefulSet) Execute() error {
	ctx := frame2.ContextOrDefault(k.Ctx)

	var err error
	k.Result, err = k.Namespace.StatefulSetInterface().Create(ctx, k.StatefulSet, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to create statefulset %q: %w", k.StatefulSet.Name, err)
	}

	return nil
}

func (k *K8SStatefulSet) Teardown() frame2.Executor {
	if !k.AutoTeardown || k.StatefulSet == nil {
		return nil
	}

	return &K8SStatefulSetRemove{
		Namespace: k.Namespace,
		Name:      k.StatefulSet.Name,
	}

}

type K8SStatefulSetRemove struct {
	Namespace *f2k8s.Namespace
	Name      string

	Ctx context.Context
}

func (k *K8SStatefulSetRemove) Execute() error {
	ctx := frame2.ContextOrDefault(k.Ctx)

	err := k.Namespace.StatefulSetInterface().Delete(ctx, k.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to remove statefulset %q: %w", k.Name, err)
	}

	return nil
}
