package disruptors

import (
	"log"

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
