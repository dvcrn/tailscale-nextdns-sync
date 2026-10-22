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

### How it works

When connected to Tailscale, if you try to access `foo.ts.mydomain.com`, tailscale can't resolve it through it's built-in MagicDNS, then matches sees the `.ts.mydomain.com` FQDN and forwards it to NextDNS.

When NextDNS tries to resolve the domain, it looks into it's internal override table and finds `foo.ts.mydomain.com` and returns the internal Tailscale IP.

Tailscale takes the IP and connects to it.

## License

MIT
