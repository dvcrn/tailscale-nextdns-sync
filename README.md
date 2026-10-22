# Tailscale -> NextDNS Sync

Small worker to sync tailnet machines to custom domain overrides in NextDNS

Eg: `foo.xxx.ts.net (100.64.0.1)` -> `foo.mydomain.ts.net` in NextDNS

## Why though?

This works as an alternative to MagicDNS. It basically recreates the entire tailnet under a custom domain you own and keeps it in sync with your machines.

You could do things like: While connected to tailscale, `foo.ts.mydomain.com` resolves over tailscale, and when not connected resolves to somewhere else (cloudflared tunnel).

## Setup

- In Tailscale, create a new OAuth credentials [here](https://console.tailscale.com/admin/settings/trust-credentials). Select `devices:core:read` as scope and write down the Client ID and Client Secret
- In NextDNS, create a new tailnet or use an existing one. Create an API Key [here](https://my.nextdns.io/account)
- Back in Tailscale, go to [DNS Settings](https://console.tailscale.com/admin/dns), add a custom nameserver, select "custom", tick `Restrict to domain`, and enter the IPv6 address of your NextDNS profile. Enter the domain you want to use in the domain field, for example `ts.mydomain.com`
- In Tailscale, add `ts.mydomain.com` as search domain

### Environment variables

| Variable | Where | Description |
| --- | --- | --- |
| `TS_OAUTH_CLIENT_ID` | secret | Tailscale OAuth client ID |
| `TS_OAUTH_CLIENT_SECRET` | secret | Tailscale OAuth client secret |
| `NEXTDNS_API_KEY` | secret | NextDNS API key |
| `NEXTDNS_PROFILE_ID` | secret | NextDNS profile that holds the rewrites, e.g. `abc123` |
| `DOMAIN` | `wrangler.jsonc` vars | Domain to sync under, e.g. `ts.mydomain.com`. Every rewrite under this domain in the profile is managed by the sync, and rewrites that don't match a current device are deleted |
| `TAILSCALE_TAILNET` | `wrangler.jsonc` vars | Tailnet ID. Optional, defaults to `-` (the OAuth client's own tailnet) |

## Running

### Deploy to Cloudflare Workers

Set your `account_id` and `DOMAIN` in `wrangler.jsonc`, then:

```sh
bun install
bunx wrangler deploy          # creates the Worker and the 5-minute cron trigger
bunx wrangler secret put TS_OAUTH_CLIENT_ID
bunx wrangler secret put TS_OAUTH_CLIENT_SECRET
bunx wrangler secret put NEXTDNS_API_KEY
bunx wrangler secret put NEXTDNS_PROFILE_ID
```

If you keep the secrets in [fnox](https://github.com/jdx/fnox), `mise run secrets` uploads all four.

### Run locally

With all variables from the table exported:

```sh
go run ./cmd/sync -n   # dry run: print planned changes
go run ./cmd/sync      # apply
```

`mise run sync` does the same with secrets loaded from fnox.

### How it works

When connected to Tailscale, if you try to access `foo.ts.mydomain.com`, tailscale can't resolve it through it's built-in MagicDNS, then matches sees the `.ts.mydomain.com` FQDN and forwards it to NextDNS.

When NextDNS tries to resolve the domain, it looks into it's internal override table and finds `foo.ts.mydomain.com` and returns the internal Tailscale IP.

Tailscale takes the IP and connects to it.

## License

MIT
