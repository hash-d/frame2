package f2skupper1

import (
	"context"
	"fmt"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2general"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
	"github.com/hash-d/frame2/pkg/frames/k8svalidate"
)

type RouterCheck struct {
	Namespace *f2k8s.Namespace

	Mode     string
	LogLevel string
	SiteName string

	Ctx context.Context

	Return *interface{}

	frame2.DefaultRunDealer
	frame2.Log
}

func (r *RouterCheck) Validate() error {

	// This will be resulting value of ConfigMap.JSON
	jsonchecks := map[string]f2general.JSON{}

	testTable := []struct {
		fieldValue string
		key        string
		matchers   []f2general.JSONMatcher
	}{
		{
			fieldValue: r.Mode,
			key:        "skrouterd.json",
			matchers: []f2general.JSONMatcher{
				{
					Expression: fmt.Sprintf("[?[0] == 'router'] |[].mode | map((&@ == '%v'), @)", r.Mode),
					Exact:      1,
				},
			},
		}, {
			fieldValue: r.LogLevel,
			key:        "skrouterd.json",
			matchers: []f2general.JSONMatcher{
				{
					Expression: fmt.Sprintf("[?[0] == 'log'] |[].enable | map((&@ == '%v'), @)", r.LogLevel),
					Exact:      1,
				},
			},
		}, {
			fieldValue: r.SiteName,
			key:        "skrouterd.json",
			matchers: []f2general.JSONMatcher{
				{
					Expression: fmt.Sprintf("[?[0] == 'router'] |[].id | map(&starts_with(@, '%v-'), @)", r.SiteName),
					Exact:      1,
				},
			},
		},
	}

	for _, i := range testTable {
		if i.fieldValue != "" {
			m := i.matchers
			entry, ok := jsonchecks[i.key]
			if !ok {
				jsonchecks[i.key] = f2general.JSON{
					Matchers: m,
				}
			} else {
				entry.Matchers = append(entry.Matchers, m...)
				jsonchecks[i.key] = entry
			}
		}
	}

	phase := frame2.Phase{
		Runner: r.GetRunner(),
		MainSteps: []frame2.Step{
			{
				Doc: "Checking contents of CM skupper-internal",
				Validator: &k8svalidate.ConfigMap{
					Namespace: r.Namespace,
					Name:      "skupper-internal",
					Ctx:       r.Ctx,

					JSON: jsonchecks,
				},
			},
		},
	}
	return phase.Run()

}
