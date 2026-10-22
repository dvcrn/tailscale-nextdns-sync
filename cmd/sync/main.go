package main

import (
	"fmt"

	"github.com/dvcrn/tailscale-nextdns-sync/internal/httpclient"
	"github.com/dvcrn/tailscale-nextdns-sync/internal/nextdns"
	"github.com/dvcrn/tailscale-nextdns-sync/internal/reconcile"
	"github.com/dvcrn/tailscale-nextdns-sync/internal/tailscale"
)

func newSyncer(getenv func(string) string) (*reconcile.Syncer, error) {
	env := map[string]string{}
	for _, k := range []string{"TS_OAUTH_CLIENT_ID", "TS_OAUTH_CLIENT_SECRET", "NEXTDNS_API_KEY", "NEXTDNS_PROFILE_ID", "DOMAIN"} {
		if env[k] = getenv(k); env[k] == "" {
			return nil, fmt.Errorf("%s is not set", k)
		}
	}
	tailnet := getenv("TAILSCALE_TAILNET")
	if tailnet == "" {
		tailnet = "-"
	}
	client := httpclient.New()
	return &reconcile.Syncer{
		Tailscale: &tailscale.Client{
			HTTP:         client,
			ClientID:     env["TS_OAUTH_CLIENT_ID"],
			ClientSecret: env["TS_OAUTH_CLIENT_SECRET"],
			Tailnet:      tailnet,
		},
		NextDNS: &nextdns.Client{
			HTTP:    client,
			APIKey:  env["NEXTDNS_API_KEY"],
			Profile: env["NEXTDNS_PROFILE_ID"],
		},
		Domain: env["DOMAIN"],
	}, nil
}
