package skupperexecute

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/utils/base"
)

// Connects two Skupper instances installed in different namespaces or clusters
//
// In practice, it does two steps: create the token, then use it to create a link
// on the other namespace
type SkupperConnect struct {
	Name string
	Cost int32
	From *base.ClusterContext
	To   *base.ClusterContext
	Ctx  context.Context

	frame2.DefaultRunDealer
	frame2.Log
}

func (sc SkupperConnect) Execute() error {
	ctx := frame2.ContextOrDefault(sc.Ctx)

	log.Printf("execute.SkupperConnect")
	var err error

	log.Printf("connecting %v to %v", sc.From.Namespace, sc.To.Namespace)

	i := rand.Intn(1000)
	secretFile := "/tmp/" + sc.To.Namespace + "_secret.yaml." + strconv.Itoa(i)
	phase := frame2.Phase{
		Runner: sc.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &TokenCreate{
					Namespace: sc.To,
					FileName:  secretFile,
				},
			},
		},
	}
	err = phase.Run()
	if err != nil {
		return fmt.Errorf("SkupperConnect failed to create token: %w", err)
	}

	var connectorCreateOpts types.ConnectorCreateOptions = types.ConnectorCreateOptions{
		SkupperNamespace: sc.From.Namespace,
		Name:             sc.Name,
		Cost:             sc.Cost,
	}
	_, err = sc.From.VanClient.ConnectorCreateFromFile(ctx, secretFile, connectorCreateOpts)
	log.Printf("SkupperConnect done")
	return err

}
