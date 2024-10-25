package frame2

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"

	"github.com/google/uuid"
)

// Populate the package variables id and smallId
func init() {

	// By default, we use the environment provided TEST_ID
	stringId := os.Getenv("TEST_ID")
	if stringId == "" {
		// If that was not provided, we create a new one, but each
		// package/binary will have its own, on a single `go test` invokation
		uu, err := uuid.NewUUID()
		if err != nil {
			uu = uuid.New()
		}
		stringId = uu.String()
	}
	hash := md5.Sum([]byte(stringId))
	asS := hex.EncodeToString(hash[:])

	// set package-level vars
	id = stringId
	shortId = asS[:3]

	log.Printf(
		"This frame2 execution is identified with UUID %v and partial MD5 %v",
		stringId,
		shortId,
	)

}
