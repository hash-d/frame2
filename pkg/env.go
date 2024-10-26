package frame2

import (
	"fmt"
	"os"
	"strconv"
)

// Frame2-specific environment variables
// TODO: move this to its own 'env' package

const (
	// This sets the 'Allow' parameter of the retry block for the final
	// validations, and needs to be an integer value.  Any validations
	// marked as final will be retried this many times at the end of the
	// test.
	ENV_FINAL_RETRY = "SKUPPER_TEST_FINAL_RETRY"
)

const (
	// If defined, both stdout and stderr of all issued skupper commands
	// will be shown on the test output, even if they did not fail
	ENV_CLI_VERBOSE_COMMANDS = "SKUPPER_TEST_CLI_VERBOSE_COMMANDS"
)

// TODO: Move all skupper-specific variables to a skupper-specific file, on a
// skupper-specific package
const (

	// Define the upgrade strategy used by the Upgrade disruptor (possibly
	// other points as well?)
	//
	// Values:
	//
	// CREATION (default): order of skupper init
	// PUB_FIRST
	// PRV_FIRST
	// PUB_ONLY
	// PRV_ONLY
	// LEAVES_FIRST
	// LEAVES_ONLY
	// CORE_FIRST
	// CORE_ONLY
	// EDGES_FIRST
	// EDGES_ONLY
	// INTERIOR_FIRST
	// INTERIOR_ONLY
	//
	// Currently, only CREATION and CREATION:INVERSE are implemented
	//
	// For any of the options, if the value ends with :INVERSE, the order
	// is inverted.  For example, ":INVERSE" or "CREATION:INVERSE" will
	// upgrade the lastly installed skupper site first; the first last.
	//
	// Valid values are of the string type TestUpgradeStrategy
	ENV_UPGRADE_STRATEGY = "SKUPPER_TEST_UPGRADE_STRATEGY"

	// A path to the Skupper binary to be used (the actual file, not just its parent directory)
	ENV_OLD_BIN = "SKUPPER_TEST_OLD_BIN"

	// The version that ENV_OLD_BIN refers to, such as 1.2 or 1.4.0-rc3
	ENV_OLD_VERSION = "SKUPPER_TEST_OLD_VERSION"

	// The expected version of the skupper binary found on the PATH
	ENV_VERSION = "SKUPPER_TEST_VERSION"

	// TODO: Change all these repeating constants by fewer constants and couple of functions:
	//
	// SkupperEnvVar(ConfigSyncImageEnvKey)
	// // This prefixes the const with SKUPPER_TEST_OLD before fetching from env
	// SkupperOldEnvVar(ConfigSyncImageEnvKey)
	// // Copy'n'paste from Skupper source code
	// ConfigSyncImageEnvKey = SKUPPER_CONFIG_SYNC_IMAGE

	// All image env variables from pkg/images/image_utils.go should be here

	EnvOldConfigSyncImageEnvKey             string = "SKUPPER_TEST_OLD_SKUPPER_CONFIG_SYNC_IMAGE"
	EnvOldConfigSyncPullPolicyEnvKey        string = "SKUPPER_TEST_OLD_SKUPPER_CONFIG_SYNC_IMAGE_PULL_POLICY"
	EnvOldFlowCollectorImageEnvKey          string = "SKUPPER_TEST_OLD_SKUPPER_FLOW_COLLECTOR_IMAGE"
	EnvOldFlowCollectorPullPolicyEnvKey     string = "SKUPPER_TEST_OLD_SKUPPER_FLOW_COLLECTOR_IMAGE_PULL_POLICY"
	EnvOldPrometheusServerImageEnvKey       string = "SKUPPER_TEST_OLD_PROMETHEUS_SERVER_IMAGE"
	EnvOldPrometheusServerPullPolicyEnvKey  string = "SKUPPER_TEST_OLD_PROMETHEUS_SERVER_IMAGE_PULL_POLICY"
	EnvOldRouterImageEnvKey                 string = "SKUPPER_TEST_OLD_QDROUTERD_IMAGE"
	EnvOldRouterPullPolicyEnvKey            string = "SKUPPER_TEST_OLD_QDROUTERD_IMAGE_PULL_POLICY"
	EnvOldServiceControllerImageEnvKey      string = "SKUPPER_TEST_OLD_SKUPPER_SERVICE_CONTROLLER_IMAGE"
	EnvOldServiceControllerPullPolicyEnvKey string = "SKUPPER_TEST_OLD_SKUPPER_SERVICE_CONTROLLER_IMAGE_PULL_POLICY"

	EnvOldSkupperImageRegistryEnvKey    string = "SKUPPER_TEST_OLD_SKUPPER_IMAGE_REGISTRY"
	EnvOldPrometheusImageRegistryEnvKey string = "SKUPPER_TEST_OLD_PROMETHEUS_IMAGE_REGISTRY"

	// Starting with 1.5
	EnvOldOauthProxyImageEnvKey            string = "SKUPPER_TEST_OLD_OAUTH_PROXY_IMAGE"
	EnvOldOauthProxyPullPolicyEnvKey       string = "SKUPPER_TEST_OLD_OAUTH_PROXY_IMAGE_PULL_POLICY"
	EnvOldControllerPodmanImageEnvKey      string = "SKUPPER_TEST_OLD_SKUPPER_CONTROLLER_PODMAN_IMAGE"
	EnvOldControllerPodmanPullPolicyEnvKey string = "SKUPPER_TEST_OLD_SKUPPER_CONTROLLER_PODMAN_IMAGE_PULL_POLICY"
	EnvOldOauthProxyRegistryEnvKey         string = "SKUPPER_TEST_OLD_OAUTH_PROXY_IMAGE_REGISTRY"
)

// final
//
// The map between the variables that indicate the image value for the old version, and the
// environment variable that actually needs to be set on the environment for that configuration
// to be effective.  Perhaps it would be simpler to just s/SKUPPER_TEST_OLD//?
var EnvOldMap = map[string]string{
	EnvOldConfigSyncImageEnvKey:             "SKUPPER_CONFIG_SYNC_IMAGE",
	EnvOldConfigSyncPullPolicyEnvKey:        "SKUPPER_CONFIG_SYNC_IMAGE_PULL_POLICY",
	EnvOldFlowCollectorImageEnvKey:          "SKUPPER_FLOW_COLLECTOR_IMAGE",
	EnvOldFlowCollectorPullPolicyEnvKey:     "SKUPPER_FLOW_COLLECTOR_IMAGE_PULL_POLICY",
	EnvOldPrometheusServerImageEnvKey:       "PROMETHEUS_SERVER_IMAGE",
	EnvOldPrometheusServerPullPolicyEnvKey:  "PROMETHEUS_SERVER_IMAGE_PULL_POLICY",
	EnvOldRouterImageEnvKey:                 "QDROUTERD_IMAGE",
	EnvOldRouterPullPolicyEnvKey:            "QDROUTERD_IMAGE_PULL_POLICY",
	EnvOldServiceControllerImageEnvKey:      "SKUPPER_SERVICE_CONTROLLER_IMAGE",
	EnvOldServiceControllerPullPolicyEnvKey: "SKUPPER_SERVICE_CONTROLLER_IMAGE_PULL_POLICY",

	EnvOldSkupperImageRegistryEnvKey:    "SKUPPER_IMAGE_REGISTRY",
	EnvOldPrometheusImageRegistryEnvKey: "PROMETHEUS_IMAGE_REGISTRY",

	// Starting with 1.5
	EnvOldOauthProxyImageEnvKey:            "OAUTH_PROXY_IMAGE",
	EnvOldOauthProxyPullPolicyEnvKey:       "OAUTH_PROXY_IMAGE_PULL_POLICY",
	EnvOldControllerPodmanImageEnvKey:      "SKUPPER_CONTROLLER_PODMAN_IMAGE",
	EnvOldControllerPodmanPullPolicyEnvKey: "SKUPPER_CONTROLLER_PODMAN_IMAGE_PULL_POLICY",
	EnvOldOauthProxyRegistryEnvKey:         "OAUTH_PROXY_IMAGE_REGISTRY",
}

func IsVerboseCommandOutput() bool {
	_, showVerbose := os.LookupEnv(ENV_CLI_VERBOSE_COMMANDS)
	return showVerbose
}

// Returns the integer value of the named variable; returns the default value if not
// defined or empty.  If there is a value and not an int, panic
//
// TODO: once this file is moved to a package 'env', the function name 'env.GetInt'
// will make more sense
func GetInt(name string, default_ int) int {
	val, found := os.LookupEnv(name)
	if !found || val == "" {
		return default_
	}
	ret, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("variable %q has non-integer value %q", name, val))
	}
	return ret

}
