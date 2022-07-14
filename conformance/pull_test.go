//go:build pull

package conformance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPullBlob(t *testing.T) {
	content := randString(t)
	digest := sha256String(content)

	nonExistentDigest := sha256String("non-existent")

	// Populate test blob.
	request{
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

	for _, r := range []request{{
		desc:       "GET blob",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/blobs/%s", env.Repo, digest),
		wantStatus: []int{http.StatusOK},
	}, {
		desc:       "HEAD blob",
		method:     http.MethodHead,
		path:       fmt.Sprintf("/v2/%s/blobs/%s", env.Repo, digest),
		wantStatus: []int{http.StatusOK},
	}, {
		desc:          "GET non-existent blob",
		method:        http.MethodGet,
		path:          fmt.Sprintf("/v2/%s/blobs/%s", env.Repo, nonExistentDigest),
		wantStatus:    []int{http.StatusNotFound},
		wantErrorCode: "BLOB_UNKNOWN",
	}, {
		desc:       "HEAD non-existent blob",
		method:     http.MethodHead,
		path:       fmt.Sprintf("/v2/%s/blobs/%s", env.Repo, nonExistentDigest),
		wantStatus: []int{http.StatusNotFound},
	}} {
		t.Run(r.desc, r.do)
	}
}

var mf = imageManifest{
	SchemaVersion: 2,
	MediaType:     "application/vnd.oci.image.manifest.v1+json",
	Config: descriptor{
		// Conformance doesn't care about the details of the config, only the mediaType.
		Digest:    sha256String(""),
		MediaType: "application/vnd.oci.image.config.v1+json",
		Size:      0,
	},
	Layers: []descriptor{{
		Digest:    sha256String("layer content"),
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Size:      int64(len("layer content")),
		Data:      []byte("layer content"),
	}, {
		Digest:    sha256String("more layer content"),
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Size:      int64(len("more layer content")),
		Data:      []byte("layer content"),
	}},
}

func TestPullManifest(t *testing.T) {
	m := imageManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			// Conformance doesn't care about the details of the config, only the mediaType.
			Digest:    sha256String(""),
			MediaType: "application/vnd.oci.image.config.v1+json",
			Size:      0,
		},
		Layers: []descriptor{{
			Digest:    sha256String("layer content"),
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Size:      int64(len("layer content")),
			Data:      []byte("layer content"),
		}, {
			Digest:    sha256String("more layer content"),
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Size:      int64(len("more layer content")),
			Data:      []byte("layer content"),
		}},
	}.string(t)
	digest := sha256String(m)
	tag := "my-tag"

	nonExistentDigest := sha256String("non-existent")
	nonExistentTag := "non-existent-tag"

	// Populate test manifest.
	request{
		method: http.MethodPut,
		path:   fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, tag),
		headers: map[string]string{
			"Content-Type":   "application/vnd.oci.image.manifest.v1+json",
			"Content-Length": fmt.Sprintf("%d", len(m)),
		},
		body:       strings.NewReader(m),
		wantStatus: []int{http.StatusCreated},
	}.do(t)

	for _, r := range []request{{
		desc:       "GET manifest by digest",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, digest),
		wantStatus: []int{http.StatusOK},
	}, {
		desc:       "GET manifest by tag",
		method:     http.MethodGet,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, tag),
		wantStatus: []int{http.StatusOK},
	}, {
		desc:       "HEAD manifest by digest",
		method:     http.MethodHead,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, digest),
		wantStatus: []int{http.StatusOK},
	}, {
		desc:       "HEAD manifest by tag",
		method:     http.MethodHead,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, tag),
		wantStatus: []int{http.StatusOK},
	}, {
		desc:          "GET non-existent manifest by digest",
		method:        http.MethodGet,
		path:          fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, nonExistentDigest),
		wantStatus:    []int{http.StatusNotFound},
		wantErrorCode: "MANIFEST_UNKNOWN",
	}, {
		desc:       "HEAD non-existent manifest by digest",
		method:     http.MethodHead,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, nonExistentDigest),
		wantStatus: []int{http.StatusNotFound},
	}, {
		desc:          "GET non-existent manifest by tag",
		method:        http.MethodGet,
		path:          fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, nonExistentTag),
		wantStatus:    []int{http.StatusNotFound},
		wantErrorCode: "MANIFEST_UNKNOWN",
	}, {
		desc:       "HEAD non-existent manifest by tag",
		method:     http.MethodHead,
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, nonExistentTag),
		wantStatus: []int{http.StatusNotFound},
	}} {
		t.Run(r.desc, r.do)
	}

	// TODO: invalid repo name should return 400, NAME_INVALID
	// TODO: unknown repo name should return 404, NAME_UNKNOWN
	// TODO: invalid tag name should return 400, NAME_INVALID(?)
}
