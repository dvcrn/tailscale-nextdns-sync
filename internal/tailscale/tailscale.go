package tailscale

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/dvcrn/tailscale-nextdns-sync/internal/httpclient"
)

const apiBase = "https://api.tailscale.com/api/v2"

// Client reads tailnet devices using an OAuth client with the devices:core:read scope.
type Client struct {
	HTTP         *http.Client
	ClientID     string
	ClientSecret string
	// Tailnet is the tailnet ID, or "-" for the OAuth client's own tailnet.
	Tailnet string
}

// Device is a tailnet node as returned by the devices API.
type Device struct {
	// Name is the MagicDNS FQDN, e.g. "comet.example.ts.net".
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	// IsExternal marks devices shared into this tailnet from another one.
	IsExternal bool `json:"isExternal"`
}

func (c *Client) Devices(ctx context.Context) ([]Device, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/tailnet/"+url.PathEscape(c.Tailnet)+"/devices", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	var out struct {
		Devices []Device `json:"devices"`
	}
	if err := httpclient.DoJSON(c.HTTP, req, &out); err != nil {
		return nil, err
	}
	return out.Devices, nil
}

func (c *Client) token(ctx context.Context) (string, error) {
	form := url.Values{
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
		"grant_type":    {"client_credentials"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := httpclient.DoJSON(c.HTTP, req, &out); err != nil {
		return "", err
	}
	return out.AccessToken, nil
}
