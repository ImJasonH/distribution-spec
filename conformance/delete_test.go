package conformance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDeleteTag(t *testing.T) {
	// Populate test tag.
	m := mf.string(t)
	tag := fmt.Sprintf("tag-%d", time.Now().Unix())
	request{
		desc:   "PUT tag",
		method: http.MethodPut,
		path:   fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, tag),
		headers: map[string]string{
			"Content-Type":   "application/vnd.oci.image.manifest.v1+json",
			"Content-Length": fmt.Sprintf("%d", len(m)),
		},
		body:       strings.NewReader(m),
		wantStatus: []int{http.StatusCreated},
	}.do(t)

	// Get the test tag.
	request{
		desc:       "GET tag",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, tag),
		wantStatus: []int{http.StatusOK},
	}.do(t)

	// Delete the test tag.
	request{
		desc:   "DELETE tag",
		method: http.MethodDelete,
		path:   fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, tag),
		wantStatus: []int{
			http.StatusAccepted,
			http.StatusMethodNotAllowed,
		},
	}.do(t)

	// Check test tag is gone.
	request{
		desc:       "GET deleted tag",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, tag),
		wantStatus: []int{http.StatusNotFound},
	}.do(t)

	// Deleted tag is not present in list.
	var tags tagListResponse
	request{
		desc:       "GET tags",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/tags/list", env.Repo),
		wantStatus: []int{http.StatusOK},
	}.do(t).unmarshal(t, tags)
	for _, found := range tags.Tags {
		if found == tag {
			t.Errorf("tag %q is still present in list after delete", tag)
		}
	}
}

func TestDeleteBlob(t *testing.T) {
	content := "blob content"
	digest := sha256String(content)

	// Populate test blob.
	request{
		desc:   "POST blob",
		method: http.MethodPost,
		path:   fmt.Sprintf("/v2/%s/blobs/uploads/", env.Repo),
		query: map[string]string{
			"digest": digest,
		},
		headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(content)),
		},
		body:       strings.NewReader(content),
		wantStatus: []int{http.StatusCreated},
	}.do(t)

	// Get the test blob.
	request{
		desc:       "GET blob",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/blobs/%s", env.Repo, digest),
		wantStatus: []int{http.StatusOK},
	}.do(t)

	// Delete the test blob.
	request{
		desc:   "DELETE blob",
		method: http.MethodDelete,
		path:   fmt.Sprintf("/v2/%s/blobs/%s", env.Repo, digest),
		wantStatus: []int{
			http.StatusAccepted,
			http.StatusMethodNotAllowed,
		},
	}.do(t)

	// Check test blob is gone.
	request{
		desc:       "GET deleted blob",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/blobs/%s", env.Repo, digest),
		wantStatus: []int{http.StatusNotFound},
	}.do(t)
}
