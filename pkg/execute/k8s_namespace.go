package execute

import (
	"fmt"
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/test/utils/base"
)

type TestRunnerCreateNamespace struct {
	Namespace    *base.ClusterContext
	AutoTearDown bool

	frame2.DefaultRunDealer
	frame2.Log
}

func (trcn TestRunnerCreateNamespace) Execute() error {
	log.Printf("TestRunnerCreateNamespace")

	log.Printf("Creating namespace %v", trcn.Namespace.Namespace)

	err := trcn.Namespace.CreateNamespace()
	if err != nil {
		return fmt.Errorf(
			"TestRunnerCreateNamespace failed to create namespace %q: %w",
			trcn.Namespace.Namespace, err,
		)
	}

	return nil
}

func (trcn TestRunnerCreateNamespace) Teardown() frame2.Executor {
	if trcn.AutoTearDown {
		return TestRunnerDeleteNamespace{
			Namespace: trcn.Namespace,
		}
	}
	return nil
}

type TestRunnerDeleteNamespace struct {
	Namespace *base.ClusterContext
}

func (trdn TestRunnerDeleteNamespace) Execute() error {
	log.Printf("Removing namespace %q", trdn.Namespace.Namespace)
	err := trdn.Namespace.DeleteNamespace()
	if err != nil {
		return fmt.Errorf(
			"TestRunnerCreateNamespace failed to delete namespace %q: %w",
			trdn.Namespace.Namespace, err,
		)
	}
	return nil
}
