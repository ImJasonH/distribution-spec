package conformance

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kelseyhightower/envconfig"
)

var env struct {
	Host      string `envconfig:"OCI_HOST" default:"http://localhost:8080"`
	Repo      string `envconfig:"OCI_REPO" default:"oci-conformance"`
	Auth      string `envconfig:"OCI_AUTH"`
	Tag       string `envconfig:"OCI_TAG" default:"my-tag"`
	MountRepo string `envconfig:"OCI_CROSSMOUNT_NAMESPACE"`
}

func init() {
	if err := envconfig.Process("oci-conformance", &env); err != nil {
		log.Fatal(err)
	}
}

type request struct {
	desc    string
	method  string
	path    string
	query   map[string]string
	headers map[string]string
	body    io.Reader

	wantStatus    []int
	wantErrorCode string
}
type response struct {
	headers http.Header
	body    string
}

func (r response) unmarshal(t *testing.T, v interface{}) {
	if err := json.NewDecoder(strings.NewReader(r.body)).Decode(&v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func (r request) do(t *testing.T) response {
	t.Helper()

	url, err := url.Parse(fmt.Sprintf("%s%s", env.Host, r.path))
	if err != nil {
		t.Fatal("parse url:", err)
	}
	q := url.Query()
	for k, v := range r.query {
		q.Add(k, v)
	}
	url.RawQuery = q.Encode()
	req, err := http.NewRequest(r.method, url.String(), r.body)
	if err != nil {
		t.Fatal("new request:", err)
	}
	for k, v := range r.headers {
		req.Header.Add(k, v)
	}
	if env.Auth != "" {
		req.Header.Add("Authorization", "Basic "+env.Auth)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("do request:", err)
	}
	t.Logf("%s %s: %d", r.method, req.URL, resp.StatusCode)

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("reading response body: %v", err)
	}
	body := string(b)

	if err := r.matchStatus(resp.StatusCode); err != nil {
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("reading response body: %v", err)
		}
		t.Logf("response body: %s", string(b))
		t.Error(err)
	}
	if err := r.matchErrorCode(resp.Body); err != nil {
		t.Error(err)
	}
	if t.Failed() {
		t.Fatal("request failed")
	}
	return response{
		headers: resp.Header,
		body:    string(body),
	}
}

func (r request) matchStatus(got int) error {
	for _, w := range r.wantStatus {
		if got == w {
			return nil
		}
	}
	return fmt.Errorf("unexpected status code: %d; want %v", got, r.wantStatus)
}

func (r request) matchErrorCode(body io.ReadCloser) error {
	if r.wantErrorCode == "" {
		return nil
	}
	defer body.Close()
	var e errorResponse
	if err := json.NewDecoder(body).Decode(&e); err != nil {
		return fmt.Errorf("unmarshal error response: %w", err)
	}
	switch len(e.Errors) {
	case 0:
		return fmt.Errorf("expected error response %q, but got no errors", r.wantErrorCode)
	case 1:
		if e.Errors[0].Code != r.wantErrorCode {
			return fmt.Errorf("unexpected error code: got %q; want %q", e.Errors[0].Code, r.wantErrorCode)
		}
		return nil
	default:
		return fmt.Errorf("unexpected number of errors: got %d; want 1", len(e.Errors))
	}
}
