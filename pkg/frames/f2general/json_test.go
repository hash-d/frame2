package f2general_test

import (
	"testing"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"gotest.tools/assert"
)

var sample = `
[
    [
        "router",
        {
            "id": "dh-1261-${HOSTNAME}",
            "mode": "edge",
            "helloMaxAgeSeconds": "3",
            "metadata": "X"
        }
    ],
    [
        "sslProfile",
        {
            "name": "skupper-amqps",
            "certFile": "/etc/skupper-router-certs/skupper-amqps/tls.crt",
            "privateKeyFile": "/etc/skupper-router-certs/skupper-amqps/tls.key",
            "caCertFile": "/etc/skupper-router-certs/skupper-amqps/ca.crt"
        }
    ],
    [
        "sslProfile",
        {
            "name": "skupper-service-client",
            "caCertFile": "/etc/skupper-router-certs/skupper-service-client/ca.crt"
        }
    ],
    [
        "listener",
        {
            "name": "amqp",
            "host": "localhost",
            "port": 5672
        }
    ],
    [
        "listener",
        {
            "name": "amqps",
            "port": 5671,
            "sslProfile": "skupper-amqps",
            "saslMechanisms": "EXTERNAL",
            "authenticatePeer": true
        }
    ],
    [
        "listener",
        {
            "name": "@9090",
            "role": "normal",
            "port": 9090,
            "http": true,
            "httpRootDir": "disabled",
            "healthz": true,
            "metrics": true
        }
    ],
    [
        "address",
        {
            "prefix": "mc",
            "distribution": "multicast"
        }
    ],
    [
        "log",
        {
            "module": "ROUTER_CORE",
            "enable": "error+"
        }
    ]
]
`

func TestJSON(t *testing.T) {

	runner := frame2.Run{
		T: t,
	}

	phase := frame2.Phase{
		Runner: &runner,
		MainSteps: []frame2.Step{
			{
				Name: "positive",
				Validators: []frame2.Validator{
					f2general.JSON{
						Data: sample,
						Matchers: []f2general.JSONMatcher{
							{
								Expression: "[?[0] == 'router'] |[].mode | map((&@ == 'edge'), @)",
								Exact:      1,
							},
						},
					},
					f2general.JSON{
						Data: sample,
						Matchers: []f2general.JSONMatcher{
							{
								Expression:  "[?[0] == 'sslProfile']",
								Exact:       2,
								NotBoolList: true,
							},
						},
					},
					f2general.JSON{
						Data: sample,
						Matchers: []f2general.JSONMatcher{
							{
								Expression:  "[?[0] == 'sslProfile']",
								Min:         2,
								Max:         2,
								NotBoolList: true,
							},
						},
					},
				},
			}, {
				Name: "negative",
				Validators: []frame2.Validator{
					f2general.JSON{
						Data: sample,
						Matchers: []f2general.JSONMatcher{
							{
								Expression: "[?[0] == 'router'] |[].mode | map((&@ == 'edge'), @)",
								Exact:      0, // Should be 1
							},
						},
					},
					f2general.JSON{
						Data: sample,
						Matchers: []f2general.JSONMatcher{
							{
								Expression: "[?[0] == 'router'] |[].mode | map((&@ == 'INVALID'), @)",
								Exact:      1, // Should be 0
							},
						},
					},
					f2general.JSON{
						Data: sample,
						Matchers: []f2general.JSONMatcher{
							{
								Expression:  "[?[0] == 'sslProfile']",
								Min:         3, // Should be 2
								NotBoolList: true,
							},
						},
					},
					f2general.JSON{
						Data: sample,
						Matchers: []f2general.JSONMatcher{
							{
								Expression:  "[?[0] == 'sslProfile']",
								Max:         1, // Should be 2
								NotBoolList: true,
							},
						},
					},
				},
				ExpectError: true,
			},
		},
	}

	assert.Assert(t, phase.Run())

}
