package f2k8s

import (
	"context"
	"fmt"
	"time"

	frame2 "github.com/hash-d/frame2/pkg"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// This is to be used as an embedded field; implement the whole of
// Namespacer here
//
// Consider this: the PodInterface below (and others) might be interesting,
// but they'd pollute the frame's namespace if embedded
//
// This is not a frame: not an executor or a validator
type Namespace struct {
	// TODO: this will not be a pointer to *base.ClusterContext,
	// as we're moving away from it; it will be its own thing
	name string

	cluster  *KubeConfig
	testBase *TestBase

	// TODO: consider removing this.  Kind (public, private, dmz) is
	// a skupper test consideration.  It might be useful elsewhere, though
	kind ClusterType
}

func (n Namespace) GetNamespaceName() string {
	return n.name
}

func (n Namespace) GetKubeConfig() *KubeConfig {
	return n.cluster
}

func (n Namespace) GetKind() ClusterType {
	return n.kind
}

// Inquires the server and returns the latest representation of the K8S Namespace
func (n Namespace) GetNamespace(ctx context.Context) (*corev1.Namespace, error) {
	ns, err := n.cluster.kubeClient.CoreV1().Namespaces().Get(
		ctx,
		n.name,
		metav1.GetOptions{},
	)
	return ns, err

}

// This is a helper to get access to the Pods API for this namespace
func (n Namespace) PodInterface() clientcorev1.PodInterface {
	return n.cluster.kubeClient.CoreV1().Pods(n.name)
}

func (n Namespace) ServiceInterface() clientcorev1.ServiceInterface {
	return n.cluster.kubeClient.CoreV1().Services(n.name)
}

func (n Namespace) DeploymentInterface() appsv1.DeploymentInterface {
	return n.cluster.kubeClient.AppsV1().Deployments(n.name)
}

func (n Namespace) KubeClient() kubernetes.Interface {
	return n.cluster.kubeClient
}

// TODO: CreateNamespaceFull (with full Namespace configuration)

// This will simply create a namespace on the given cluster, with the
// requested name.
//
// For most uses, you may want to use CreateNamespaceTestBase, instead.
//
// Created namespaces will be labeled with frame2.id
type CreateNamespaceRaw struct {
	Name    string
	Cluster *KubeConfig

	AutoTearDown bool

	Annotations map[string]string
	Labels      map[string]string

	frame2.DefaultRunDealer
	frame2.Log

	Return corev1.Namespace
}

func (c *CreateNamespaceRaw) Execute() error {
	c.Log.Printf("Creating k8s namespace %q", c.Name)

	err := ConnectInitial()
	if err != nil {
		return fmt.Errorf("failed to create initial connection to the clusters: %w", err)
	}

	labels := c.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labels["frame2.id"] = frame2.GetId()
	labels["frame2.shortid"] = frame2.GetShortId()

	ns, err := c.Cluster.GetKubeClient().CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        c.Name,
				Labels:      labels,
				Annotations: c.Annotations,
			},
		},
		metav1.CreateOptions{},
	)
	c.Return = *ns
	return err
}

func (c *CreateNamespaceRaw) Teardown() frame2.Executor {
	if c.AutoTearDown {
		return &DeleteNamespaceRaw{
			Namespace: c.Name,
			Cluster:   c.Cluster,
		}
	}
	return nil
}

// This creates a K8S namespace, based on a TestBase; the namespace name
// will include the provided Id, but it will also contain other information
// (such as the TestBase Id, a sequence number, etc)
//
// The namespace also receives a label frame2.id, with which it is easy to
// identify (and possibly modify or remove) any namespaces created by a test
// run.
type CreateNamespaceTestBase struct {
	Id       string
	TestBase *TestBase
	Kind     ClusterType

	AutoTearDown bool

	// Annotations to be applied after the namespace creation
	Annotations map[string]string

	// Labels to be applied after the namespace creation
	Labels map[string]string

	frame2.DefaultRunDealer
	frame2.Log

	Return Namespace
}

func (c *CreateNamespaceTestBase) Execute() (err error) {

	c.Log.Printf("Id: %q, Kind: %q", c.Id, c.Kind)

	f2ns := Namespace{
		testBase: c.TestBase,
		kind:     c.Kind,
	}

	labels := c.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labels["frame2.testbase"] = c.TestBase.namespaceId
	if c.Id != "" {
		labels["frame2.ns.id"] = c.Id
	}

	raw := CreateNamespaceRaw{
		Labels:       labels,
		Annotations:  c.Annotations,
		AutoTearDown: c.AutoTearDown,
	}

	// Get the name and cluster for the namespace to be created,
	// and defer .Add() to release the lock, regarless of result
	name, cluster := c.TestBase.Next(c.Kind, c.Id)
	defer c.TestBase.Add(&f2ns, err)

	f2ns.name = name
	raw.Name = name
	raw.Cluster = cluster
	f2ns.cluster = cluster

	c.Return = f2ns

	// Actually create the namespace
	phase := frame2.Phase{
		Runner: c.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Modify: &raw,
			},
		},
	}
	err = phase.Run()

	return

}

func (c *CreateNamespaceTestBase) Teardown() frame2.Executor {
	if c.AutoTearDown {
		return &DeleteNamespaceTestBase{
			Namespace: &c.Return,
		}
	}
	return nil
}

type DeleteNamespaceTestBase struct {
	Namespace *Namespace

	// A wait duration of zero will use the default wait
	// for namespace removals; if you want to not wait at all
	// use a very small duration (such as time.Nanosecond)
	Wait time.Duration

	frame2.DefaultRunDealer
	frame2.Log
}

func (d *DeleteNamespaceTestBase) Execute() error {
	if d.Namespace == nil {
		return fmt.Errorf("cannot remove nil namespace")
	}
	phase := frame2.Phase{
		Runner: d.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Modify: &DeleteNamespaceRaw{
					Namespace: d.Namespace.name,
					Cluster:   d.Namespace.cluster,
					Wait:      d.Wait,
				},
			},
		},
	}
	return phase.Run()
}

// This is a direct call to K8S to remove a namespace.  For namespaces created
// with CreateNamespaceTestBase, you want to use DeleteNamespaceTestBase,
// instead
type DeleteNamespaceRaw struct {
	Namespace string
	Cluster   *KubeConfig

	// A wait duration of zero will use the default wait
	// for namespace removals; if you want to not wait at all
	// use a very small duration (such as time.Nanosecond)
	Wait time.Duration

	frame2.DefaultRunDealer
	frame2.Log
}

func (d *DeleteNamespaceRaw) Execute() error {
	d.Log.Printf("Removing namespace %q", d.Namespace)

	// Create a namespace informer
	done := make(chan struct{})
	factory := informers.NewSharedInformerFactory(d.Cluster.kubeClient, 0)
	nsInformer := factory.Core().V1().Namespaces().Informer()
	nsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			ns, _ := cache.MetaNamespaceKeyFunc(obj)
			// when requested namespace has been deleted, close the done channel
			if ns == d.Namespace {
				close(done)
			}
		},
	})

	stop := make(chan struct{})
	go nsInformer.Run(stop)

	// Delete the ns
	err := d.Cluster.kubeClient.CoreV1().Namespaces().Delete(
		// TODO: change this with the framework's provided Context
		context.TODO(),
		d.Namespace,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return err
	}

	wait := d.Wait
	if wait == 0 {
		// TODO: move this to a constant, or a configurable value
		wait = time.Minute * 2
	}

	// Wait for informer to be done or a timeout
	timeout := time.After(wait)
	select {
	case <-timeout:
		err = fmt.Errorf("timed out waiting on namespace %q to be deleted after %v", d.Namespace, wait)
	case <-done:
		break
	}

	// stop informer
	close(stop)

	return err
}
