package f2skupper1

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

// Connects two Skupper instances installed in different namespaces or clusters
//
// In practice, it does two steps: create the token, then use it to create a
// link on the other namespace
//
// This _does not_ implement SkupperVersioner: as it calls TokenCreate and
// LinkCreate, the logic for version-specific behavior should be on them.  For
// all Connect knows, the two sites could even be on different versions.  Keep
// Connect simple.
type Connect struct {
	From *f2k8s.Namespace
	To   *f2k8s.Namespace

	SecretName string
	Expiry     string
	Password   string
	TokenType  string
	Uses       string

	LinkName string
	Cost     string

	Ctx context.Context

	frame2.DefaultRunDealer
	frame2.Log
}

func (sc Connect) Execute() error {
	// ctx := frame2.ContextOrDefault(sc.Ctx)

	log.Printf("execute.SkupperConnect")

	log.Printf("connecting %v to %v", sc.From.GetNamespaceName(), sc.To.GetNamespaceName())

	i := rand.Intn(1000)
	secretFile := "/tmp/" + sc.To.GetNamespaceName() + "_secret.yaml." + strconv.Itoa(i)
	phase := frame2.Phase{
		Runner: sc.Runner,
		Doc:    fmt.Sprintf("Connect skupper from namespace %q to %q", sc.From.GetNamespaceName(), sc.To.GetNamespaceName()),
		MainSteps: []frame2.Step{
			{
				Modify: &TokenCreate{
					Namespace: sc.To,
					FileName:  secretFile,
					Expiry:    sc.Expiry,
					Name:      sc.SecretName,
					Password:  sc.Password,
					TokenType: sc.TokenType,
					Uses:      sc.Uses,
				},
			}, {
				Modify: &LinkCreate{
					Namespace: sc.From,
					File:      secretFile,
					Name:      sc.LinkName,
					Cost:      sc.Cost,
				},
			},
		},
	}
	return phase.Run()

}
