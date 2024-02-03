package util

import (
	"encoding/base64"
	"fmt"
	"time"
)

// NewEtag generates a new Etag for response caching
func NewEtag() string {
	now := fmt.Sprintf("%d", time.Now().Unix())
	etag := base64.StdEncoding.EncodeToString([]byte(now))

	return etag
}
