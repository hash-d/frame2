package f2k8s

import (
	"fmt"
	"log"
	"sync"

	frame2 "github.com/hash-d/frame2/pkg"
	openshiftapps "github.com/openshift/client-go/apps/clientset/versioned"

	routev1client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clusters = []*KubeConfig{}
var domainClusters = map[ClusterType][]*KubeConfig{}
var namedClusters = map[string]*KubeConfig{}

// Creates the KubeConfig via NewKubeConfig with the provided data,
// and includes it on the private package vars clusters, domainClusters
// and namedClusters
func addCluster(domain ClusterType, name, file string) error {
	if _, ok := namedClusters[name]; ok {
		return fmt.Errorf("a cluster named %q already exists", name)
	}
	kubeconfig, err := NewKubeConfig(name, file)
	if err != nil {
		return err
	}
	clusters = append(clusters, kubeconfig)
	namedClusters[name] = kubeconfig
	domainClusters[domain] = append(domainClusters[domain], kubeconfig)
	return nil
}

var once = sync.Once{}

// Connect to the clusters declared with -kube or on KUBECONFIG
//
// This can be called many times, but will be only executed once.
func ConnectInitial() error {
	var err error
	once.Do(func() {
		log.Printf("Connecting the the clusters...")
		if len(contexts) == 0 {
			// No kubeconfigs parsed from the command line; we'll
			// use $KUBECONFIG or ~/.kube/config (per
			// k8s.io/client-go/tools/clientcmd/ClientConfigLoadingRules)
			// and call it simply "pub", as a pub ClusterType
			err = addCluster("pub", "pub", "")
			return
		}
		for domain, files := range contexts {
			if len(files) == 1 {
				err = addCluster(domain, string(domain), files[0])
				if err != nil {
					return
				}
				continue
			}
			for i, f := range files {
				err = addCluster(domain, fmt.Sprintf("%s%d", domain, i), f)
				if err != nil {
					return
				}
			}
		}
	})
	return err
}

// TODO: make this an Executor?
type KubeConfig struct {
	name            string
	kubeConfigPath  string
	restConfig      *rest.Config
	kubeClient      *kubernetes.Clientset
	routeClient     *routev1client.RouteV1Client
	ocAppsClient    *openshiftapps.Clientset
	discoveryClient *discovery.DiscoveryClient
	dynamicClient   dynamic.Interface

	Log *frame2.Log
}

func (k KubeConfig) GetKubeconfigFile() string {
	return k.kubeConfigPath
}

func NewKubeConfig(name, path string) (*KubeConfig, error) {
	k := KubeConfig{
		kubeConfigPath: path,
		name:           name,
	}
	err := k.connect()

	return &k, err
}

func (k KubeConfig) GetName() string {
	return k.name
}

// Returns the KubeClient for interacting with the cluster defined on this
// KubeConfig
func (k KubeConfig) GetKubeClient() *kubernetes.Clientset {
	return k.kubeClient
}

func (k KubeConfig) GetRouteClient() *routev1client.RouteV1Client {
	return k.routeClient
}

func (k KubeConfig) GetOcAppsClient() *openshiftapps.Clientset {
	return k.ocAppsClient
}

func (k KubeConfig) GetDiscoveryClient() *discovery.DiscoveryClient {
	return k.discoveryClient
}

func (k KubeConfig) GetDynamicClient() dynamic.Interface {
	return k.dynamicClient
}

func (k KubeConfig) GetRestConfig() *rest.Config {
	return k.restConfig
}

// This uses kubernetest.NewForConfig() to create the kubeClient for this
// KubeConfig, and all other clients
func (k *KubeConfig) connect() error {
	// TODO: improve error handling: wrap errors with additional info
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if k.kubeConfigPath != "" {
		loadingRules = &clientcmd.ClientConfigLoadingRules{
			ExplicitPath: k.kubeConfigPath,
		}
	}
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)
	rawConfig, err := kubeconfig.RawConfig()
	if err != nil {
		return err
	}
	server := ""
	if cluster, ok := rawConfig.Clusters[rawConfig.CurrentContext]; ok {
		server = cluster.Server
	}
	k.Log.Printf(
		"KubeConfig: Connecting %q to server %q (%q)",
		k.name,
		server,
		k.kubeConfigPath,
	)
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return err
	}
	restconfig.ContentConfig.GroupVersion = &schema.GroupVersion{Version: "v1"}
	restconfig.APIPath = "/api"
	restconfig.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: codecs}
	k.restConfig = restconfig
	k.kubeClient, err = kubernetes.NewForConfig(restconfig)
	if err != nil {
		return err
	}
	k.discoveryClient, err = discovery.NewDiscoveryClientForConfig(restconfig)
	resources, err := k.discoveryClient.ServerResourcesForGroupVersion("route.openshift.io/v1")
	if err == nil && len(resources.APIResources) > 0 {
		k.routeClient, err = routev1client.NewForConfig(restconfig)
		if err != nil {
			return err
		}
	}
	resources, err = k.discoveryClient.ServerResourcesForGroupVersion("apps.openshift.io/v1")
	if err == nil && len(resources.APIResources) > 0 {
		k.ocAppsClient, err = openshiftapps.NewForConfig(restconfig)
		if err != nil {
			return err
		}
	}

	k.dynamicClient, err = dynamic.NewForConfig(restconfig)
	if err != nil {
		return err
	}

	return nil
}
