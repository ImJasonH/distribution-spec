package conformance

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

func sha256String(s string) string { return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(s))) }

type imageManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        descriptor   `json:"config"`
	Layers        []descriptor `json:"layers"`
}

func (im imageManifest) string(t *testing.T) string {
	b, err := json.Marshal(im)
	if err != nil {
		t.Fatal("marshal image manifest:", err)
	}
	return string(b)
}

type descriptor struct {
	Digest    string `json:"digest"`
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Data      []byte `json:"data"`
}

type errorResponse struct {
	Errors []singleError `json:"errors"`
}
type singleError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Conformance doesn't care about error details.
}
