package nextdns

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestRequestBody(t *testing.T) {
	var got *http.Request
	c := &Client{
		HTTP: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			got = req
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"data":[]}`))}, nil
		})},
		APIKey:  "key",
		Profile: "abc123",
	}
	ctx := context.Background()

	if _, err := c.Rewrites(ctx); err != nil {
		t.Fatal(err)
	}
	if got.Body != nil {
		t.Errorf("GET body = %#v, want nil", got.Body)
	}

	if err := c.Delete(ctx, "id1"); err != nil {
		t.Fatal(err)
	}
	if got.Body != nil {
		t.Errorf("DELETE body = %#v, want nil", got.Body)
	}

	if err := c.Create(ctx, "comet.ts.d.pn", "100.64.0.1"); err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(got.Body)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"name":"comet.ts.d.pn","content":"100.64.0.1"}`; string(body) != want {
		t.Errorf("POST body = %s, want %s", body, want)
	}
}
