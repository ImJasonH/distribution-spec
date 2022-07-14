// +pull || all
package conformance

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPullBlob(t *testing.T) {
	content := "blob content"
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
	}, {
		Digest:    sha256String("more layer content"),
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Size:      int64(len("more layer content")),
	}},
}

func TestPullManifest(t *testing.T) {
	m := mf.string(t)
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
		path:       fmt.Sprintf("/v2/%s/manifests/%s", env.Repo, digest),
		wantStatus: []int{http.StatusOK},
	}, {
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

/*
var test01Pull = func() {
	g.Context(titlePull, func() {

		var tag string

		g.Context("Setup", func() {
			g.Specify("Populate registry with test blob", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, _ := client.Do(req)
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", configs[0].Digest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", configs[0].ContentLength).
					SetBody(configs[0].Content)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test layer", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, _ := client.Do(req)
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", layerBlobDigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", layerBlobContentLength).
					SetBody(layerBlobData)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test manifest", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				tag := testTagName
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tag)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifests[0].Content)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Get the name of a tag", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, _ := client.Do(req)
				tag = getTagNameFromResponse(resp)

				// attempt to forcibly overwrite this tag with the unique manifest for this run
				req = client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tag)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifests[0].Content)
				_, _ = client.Do(req)
			})

			g.Specify("Get tag name from environment", func() {
				SkipIfDisabled(pull)
				RunOnlyIfNot(runPullSetup)
				tag = os.Getenv(envVarTagName)
			})
		})

		g.Context("Pull blobs", func() {
			g.Specify("HEAD request to nonexistent blob should result in 404 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("HEAD request to existing blob should yield 200", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(configs[0].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(configs[0].Digest))
				}
			})

			g.Specify("GET nonexistent blob should result in 404 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>",
					reggie.WithDigest(dummyDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to existing blob URL should yield 200", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configs[0].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Pull manifests", func() {
			g.Specify("HEAD request to nonexistent manifest should return 404", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(nonexistentManifest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("HEAD request to manifest path (digest) should yield 200 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.HEAD, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[0].Digest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(manifests[0].Digest))
				}
			})

			g.Specify("HEAD request to manifest path (tag) should yield 200 response", func() {
				SkipIfDisabled(pull)
				Expect(tag).ToNot(BeEmpty())
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>", reggie.WithReference(tag)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				if h := resp.Header().Get("Docker-Content-Digest"); h != "" {
					Expect(h).To(Equal(manifests[0].Digest))
				}
			})

			g.Specify("GET nonexistent manifest should return 404", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(nonexistentManifest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))
			})

			g.Specify("GET request to manifest path (digest) should yield 200 response", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[0].Digest)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})

			g.Specify("GET request to manifest path (tag) should yield 200 response", func() {
				SkipIfDisabled(pull)
				Expect(tag).ToNot(BeEmpty())
				req := client.NewRequest(reggie.GET, "/v2/<name>/manifests/<reference>", reggie.WithReference(tag)).
					SetHeader("Accept", "application/vnd.oci.image.manifest.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			})
		})

		g.Context("Error codes", func() {
			g.Specify("400 response body should contain OCI-conforming JSON message", func() {
				SkipIfDisabled(pull)
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference("sha256:totallywrong")).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(invalidManifestContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					Equal(http.StatusBadRequest),
					Equal(http.StatusNotFound)))
				if resp.StatusCode() == http.StatusBadRequest {
					errorResponses, err := resp.Errors()
					Expect(err).To(BeNil())

					Expect(errorResponses).ToNot(BeEmpty())
					Expect(errorCodes).To(ContainElement(errorResponses[0].Code))
				}
			})
		})

		g.Context("Teardown", func() {
			if deleteManifestBeforeBlobs {
				g.Specify("Delete manifest created in setup", func() {
					SkipIfDisabled(pull)
					RunOnlyIf(runPullSetup)
					req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[0].Digest))
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode()).To(SatisfyAny(
						SatisfyAll(
							BeNumerically(">=", 200),
							BeNumerically("<", 300),
						),
						Equal(http.StatusMethodNotAllowed),
					))
				})
			}

			g.Specify("Delete config blob created in setup", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configs[0].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300),
					),
					Equal(http.StatusMethodNotAllowed),
				))
			})

			g.Specify("Delete layer blob created in setup", func() {
				SkipIfDisabled(pull)
				RunOnlyIf(runPullSetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(layerBlobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300),
					),
					Equal(http.StatusMethodNotAllowed),
				))
			})

			if !deleteManifestBeforeBlobs {
				g.Specify("Delete manifest created in setup", func() {
					SkipIfDisabled(pull)
					RunOnlyIf(runPullSetup)
					req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(manifests[0].Digest))
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode()).To(SatisfyAny(
						SatisfyAll(
							BeNumerically(">=", 200),
							BeNumerically("<", 300),
						),
						Equal(http.StatusMethodNotAllowed),
					))
				})
			}
		})
	})
}
*/
