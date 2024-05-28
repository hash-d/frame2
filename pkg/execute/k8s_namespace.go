package execute

import (
	"fmt"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OBSOLETE: use f2k8s.CreateNamespaceTestBase, instead
type testRunnerCreateNamespace struct {
	Namespace    *base.ClusterContext
	AutoTearDown bool

	// Annotations to be applied after the namespace creation
	Annotations map[string]string

	// Labels to be applied after the namespace creation
	Labels map[string]string

	frame2.DefaultRunDealer
	frame2.Log
}

func (cn testRunnerCreateNamespace) Execute() error {
	log.Printf("testRunnerCreateNamespace")

	log.Printf("Creating namespace %v", cn.Namespace.Namespace)

	err := cn.Namespace.CreateNamespace()
	if err != nil {
		return fmt.Errorf(
			"testRunnerCreateNamespace failed to create namespace %q: %w",
			cn.Namespace.Namespace, err,
		)
	}

	if cn.Annotations != nil || cn.Labels != nil {
		n, err := cn.Namespace.VanClient.KubeClient.CoreV1().Namespaces().Get(
			cn.Runner.GetContext(),
			cn.Namespace.Namespace,
			metav1.GetOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to get just-created namespace %s", cn.Namespace.Namespace)
		}
		if n.Labels == nil {
			n.Labels = make(map[string]string)
		}
		// We merge anything already there by overwritting any existing keys
		for k, v := range cn.Labels {
			n.Labels[k] = v
		}
		if n.Annotations == nil {
			n.Annotations = make(map[string]string)
		}
		// We merge anything already there by overwritting any existing keys
		for k, v := range cn.Annotations {
			n.Annotations[k] = v
		}
		_, err = cn.Namespace.VanClient.KubeClient.CoreV1().Namespaces().Update(
			cn.Runner.GetContext(),
			n,
			metav1.UpdateOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to update just-created namespace %s", cn.Namespace.Namespace)
		}
	}

	return nil
}

func (trcn testRunnerCreateNamespace) Teardown() frame2.Executor {
	if trcn.AutoTearDown {
		return testRunnerDeleteNamespace{
			Namespace: trcn.Namespace,
		}
	}
	return nil
}

type testRunnerDeleteNamespace struct {
	Namespace *base.ClusterContext
}

func (trdn testRunnerDeleteNamespace) Execute() error {
	log.Printf("Removing namespace %q", trdn.Namespace.Namespace)
	err := trdn.Namespace.DeleteNamespace()
	if err != nil {
		return fmt.Errorf(
			"testRunnerCreateNamespace failed to delete namespace %q: %w",
			trdn.Namespace.Namespace, err,
		)
	}
	return nil
}
