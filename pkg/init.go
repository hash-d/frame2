package frame2

import (
	"crypto/md5"
	"encoding/hex"
	"log"

	"github.com/google/uuid"
)

// Populate the package variables id and smallId
func init() {

	_id, err := uuid.NewUUID()
	if err != nil {
		_id = uuid.New()
	}
	hash := md5.Sum([]byte(_id.String()))
	asS := hex.EncodeToString(hash[:])

	id = _id.String()
	shortId = asS[:3]

	log.Printf(
		"This frame2 execution is identified with UUID %v and partial MD5 %v",
		id,
		shortId,
	)

}
