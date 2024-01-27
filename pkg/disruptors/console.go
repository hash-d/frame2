package disruptors

import (
	"fmt"
	"log"
	"strings"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/skupperexecute"
)

// Any skupper init runs will be overridden to not use the
// console
type NoConsole struct{}

func (n NoConsole) DisruptorEnvValue() string {
	return "NO_CONSOLE"
}

func (u *NoConsole) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if mod, ok := step.Modify.(*skupperexecute.CliSkupperInstall); ok {
		mod.EnableConsole = false
		log.Printf("NO_CONSOLE: %v", mod.Namespace.Namespace)
	}
}

type ConsoleOnAll struct{}

func (c ConsoleOnAll) DisruptorEnvValue() string {
	return "CONSOLE_ON_ALL"

}

func (c *ConsoleOnAll) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if mod, ok := step.Modify.(*skupperexecute.CliSkupperInstall); ok {
		mod.EnableConsole = true
		log.Printf("CONSOLE_ON_ALL: %v", mod.Namespace.Namespace)
	}
}

// TODO move this to its own file
type NoFlowCollector struct{}

func (n NoFlowCollector) DisruptorEnvValue() string {
	return "NO_FLOW_COLLECTOR"
}

func (u *NoFlowCollector) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if mod, ok := step.Modify.(*skupperexecute.CliSkupperInstall); ok {
		mod.EnableFlowCollector = false
		log.Printf("NO_FLOW_COLLECTOR: %v", mod.Namespace.Namespace)
	}
}

type FlowCollectorOnAll struct{}

func (f FlowCollectorOnAll) DisruptorEnvValue() string {
	return "FLOW_COLLECTOR_ON_ALL"

}

func (f FlowCollectorOnAll) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if mod, ok := step.Modify.(*skupperexecute.CliSkupperInstall); ok {
		mod.EnableFlowCollector = true
		log.Printf("FLOW_COLLECTOR_ON_ALL: %v", mod.Namespace.Namespace)
	}
}

// Overwrite the console authentication used
//
// Configure with the keywords mode, user and password, separated by commas.
//
// eg: CONSOLE_AUTH:mode=internal,user=asdf,password=foo
type ConsoleAuth struct {
	Mode     string
	User     string
	Password string
}

func (c ConsoleAuth) DisruptorEnvValue() string {
	return "CONSOLE_AUTH"
}

func (c *ConsoleAuth) Inspect(step *frame2.Step, phase *frame2.Phase) {
	if mod, ok := step.Modify.(*skupperexecute.CliSkupperInstall); ok {
		mod.ConsoleAuth = c.Mode
		mod.ConsoleUser = c.User
		mod.ConsolePassword = c.Password
		log.Printf("CONSOLE_AUTH: %v", mod.Namespace.Namespace)
	}
}

func (c *ConsoleAuth) Configure(config string) error {
	items := strings.Split(config, ",")
	for _, i := range items {
		definition := strings.Split(i, "=")
		if len(definition) != 2 {
			return fmt.Errorf("%q is not a valid ConsoleAuth configuration", i)
		}
		k, v := definition[0], definition[1]
		switch k {
		case "mode":
			c.Mode = v
		case "user":
			c.User = v
		case "password":
			c.Password = v
		default:
			return fmt.Errorf("The key %q is not valid for ConsoleAuth configuration", k)
		}
	}
	return nil
}
