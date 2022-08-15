package conformance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type tagListResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func TestContentDiscovery(t *testing.T) {
	numTags := 4
	wantTags := make([]string, 0, numTags)
	m := mf.string(t)

	// Populate test tags.
	for i := 0; i < numTags; i++ {
		tag := fmt.Sprintf("tag-%d", i)
		wantTags = append(wantTags, tag)
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
	}

	t.Run("get tags", func(t *testing.T) {
		var tags tagListResponse
		request{
			method:     http.MethodGet,
			path:       fmt.Sprintf("/v2/%s/tags/list", env.Repo),
			wantStatus: []int{http.StatusOK},
		}.do(t).unmarshal(t, &tags)
		want := tagListResponse{
			Name: env.Repo,
			Tags: wantTags,
		}
		if d := cmp.Diff(want, tags); d != "" {
			t.Errorf("tag list diff (-want,+got):\n%s", d)
		}
	})

	// Limit tags using `n` query parameter.
	t.Run("limit tags", func(t *testing.T) {
		n := numTags / 2
		var tags tagListResponse
		request{
			method:     http.MethodGet,
			path:       fmt.Sprintf("/v2/%s/tags/list?n=%d", env.Repo, n),
			wantStatus: []int{http.StatusOK},
		}.do(t).unmarshal(t, &tags)
		want := tagListResponse{
			Name: env.Repo,
			Tags: wantTags[:n],
		}
		if d := cmp.Diff(want, tags); d != "" {
			t.Errorf("tag list diff (-want,+got):\n%s", d)
		}
	})

	// Start tag list using `last` query parameter.
	t.Run("start tag list", func(t *testing.T) {
		last := wantTags[numTags/2]
		var tags tagListResponse
		request{
			method:     http.MethodGet,
			path:       fmt.Sprintf("/v2/%s/tags/list?last=%s", env.Repo, last),
			wantStatus: []int{http.StatusOK},
		}.do(t).unmarshal(t, &tags)
		want := tagListResponse{
			Name: env.Repo,
			Tags: wantTags[(numTags/2)+1:],
		}
		if d := cmp.Diff(want, tags); d != "" {
			t.Errorf("tag list diff (-want,+got):\n%s", d)
		}
	})
}
