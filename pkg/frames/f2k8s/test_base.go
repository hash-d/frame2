package f2k8s

import (
	"fmt"
	"sync"

	frame2 "github.com/hash-d/frame2/pkg"
)

// This contains a list of namespaces to be used by the test
//
// Use NewTestBase() as a constructor.
//
// TODO description
type TestBase struct {
	allNamespaces []*Namespace
	namespaces    map[string]*Namespace
	domainById    map[ClusterType][]*Namespace

	lock            sync.Mutex
	receivedKind    ClusterType
	providedName    string
	providedCluster *KubeConfig

	// An Id that will be part of the name of all namespaces created from
	// this TestBase
	namespaceId string
}

func NewTestBase(id string) *TestBase {
	return &TestBase{
		namespaceId: id,
		namespaces:  make(map[string]*Namespace),
		domainById:  make(map[ClusterType][]*Namespace),
	}
}

func (t *TestBase) GetAllNamespaces() []*Namespace {
	return t.allNamespaces
}

// Return the named namespace, if it was created by this TestBase; nil otherwise
func (t *TestBase) GetNamespace(name string) *Namespace {
	return t.namespaces[name]
}

// Return all namespaces of a given domain, such as "prv" or "dmz"
func (t *TestBase) GetDomainNamespaces(domain ClusterType) []*Namespace {
	return t.domainById[domain]
}

// Returns the name and cluster for the next namespace of `kind` in this TestBase,
// with the expectation that the caller will then immediately create the namespace.
//
// Right after this call, the caller _must_ defer t.Add(), as t.Next() locks a
// mutex and t.Add() unlocks it — even if the namespace creation operation was
// not successful.  Failing to do so may create deadlocks
//
// If the optional suffix is provided, it will be appended to the end of the name.
//
// Panics if kind is empty.  If there are not clusters of the corresponding kind,
// however, it will use the "pub" list to return a cluster.  This allows, for example,
// tests to request 'pub' and 'prv' namespaces, but run on a single cluster.
func (t *TestBase) Next(kind ClusterType, suffix string) (name string, cluster *KubeConfig) {
	t.lock.Lock()
	if kind == "" {
		panic("kind must be provided")
	}
	if t.receivedKind != "" || t.providedName != "" || t.providedCluster != nil {
		// This is an assertion; it should never happen
		panic(
			fmt.Sprintf(
				"non-empty receiveKind (%q), providedName (%q) or providedCluster(%v) at beginning of Next()",
				t.receivedKind,
				t.providedName,
				t.providedCluster,
			),
		)
	}
	actualKind := kind
	clusterList := domainClusters[kind]
	if len(clusterList) == 0 {
		actualKind = Public
		clusterList = domainClusters[actualKind]
	}
	nsList := t.domainById[actualKind]
	nextId := len(nsList)
	clusterSelect := nextId % len(clusterList)
	cluster = clusterList[clusterSelect]

	name = fmt.Sprintf(
		"%s-%s-%s-%d",
		kind,
		frame2.GetShortId(),
		t.namespaceId,
		nextId,
	)
	if suffix != "" {
		name += "-" + suffix
	}

	t.receivedKind = kind
	t.providedName = name
	t.providedCluster = cluster

	return
}

// This must be called only and always right after Next(); it will panic otherwise.
//
// Adds a namespace to the lists (if receivedErr is nil)
func (t *TestBase) Add(ns *Namespace, receivedErr error) {
	defer t.lock.Unlock()

	if t.receivedKind == "" || t.providedName == "" || t.providedCluster == nil {
		panic("Add() must be called right after Next()")
	}

	if receivedErr == nil {

		t.allNamespaces = append(t.allNamespaces, ns)
		t.namespaces[t.providedName] = ns
		t.domainById[t.receivedKind] = append(t.domainById[t.receivedKind], ns)
	}

	t.receivedKind = ""
	t.providedName = ""
	t.providedCluster = nil
}
