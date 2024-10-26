package f2skconst

// Constants on this file are copied straight from skupper code
//
// Perhaps those would be best served as functions?  Or in the future they
// may be versioned, to reflect changes on Skupper itself.

// api/types/types.go

// Transport constants
const (
	TransportDeploymentName string = "skupper-router"
	TransportComponentName  string = "router"
	ConfigSyncContainerName string = "config-sync"
)

// Controller and Collector constants
const (
	ControllerDeploymentName   string = "skupper-service-controller"
	ControllerComponentName    string = "service-controller"
	ControllerContainerName    string = "service-controller"
	FlowCollectorContainerName string = "flow-collector"
	PrometheusDeploymentName   string = "skupper-prometheus"
	PrometheusContainerName    string = "prometheus-server"
)

// Skupper qualifiers
const (
	BaseQualifier       string = "skupper.io"
	ComponentAnnotation string = BaseQualifier + "/component"
)
