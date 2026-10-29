package nextdns

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/dvcrn/tailscale-nextdns-sync/internal/httpclient"
)

const apiBase = "https://api.nextdns.io"

// Client manages the rewrites of a single NextDNS profile.
type Client struct {
	HTTP    *http.Client
	APIKey  string
	Profile string
}

// Rewrite is a NextDNS rewrite; the record type is derived from Content.
type Rewrite struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (c *Client) Rewrites(ctx context.Context) ([]Rewrite, error) {
	var out struct {
		Data []Rewrite `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, "", nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) Create(ctx context.Context, name, content string) error {
	body, err := json.Marshal(Rewrite{Name: name, Content: content})
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPost, "", body, nil)
}

func (c *Client) Delete(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/"+url.PathEscape(id), nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body []byte, out any) error {
	// workers-go sends any non-nil Body, even http.NoBody, and fetch rejects a GET with a body.
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, apiBase+"/profiles/"+url.PathEscape(c.Profile)+"/rewrites"+path, r)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return httpclient.DoJSON(c.HTTP, req, out)
}
