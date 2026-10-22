package reconcile

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dvcrn/tailscale-nextdns-sync/internal/nextdns"
	"github.com/dvcrn/tailscale-nextdns-sync/internal/tailscale"
)

// Record is one desired rewrite: a device name under the managed domain and one of its addresses.
type Record struct {
	Name    string
	Content string
}

// Desired returns one record per address of every device in the tailnet, named <label>.<domain>.
func Desired(devices []tailscale.Device, domain string) []Record {
	seen := map[Record]bool{}
	var out []Record
	for _, d := range devices {
		label, _, _ := strings.Cut(d.Name, ".")
		if d.IsExternal || label == "" {
			continue
		}
		for _, addr := range d.Addresses {
			r := Record{Name: strings.ToLower(label + "." + domain), Content: addr}
			if !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	return out
}

// Plan diffs desired against current. Every rewrite under domain is owned by the sync, so
// rewrites there that are not desired get deleted; rewrites outside domain are never touched.
func Plan(desired []Record, current []nextdns.Rewrite, domain string) (create []Record, del []nextdns.Rewrite) {
	want := map[Record]bool{}
	for _, r := range desired {
		want[r] = true
	}
	suffix := "." + strings.ToLower(domain)
	have := map[Record]bool{}
	for _, rw := range current {
		r := Record{Name: strings.ToLower(rw.Name), Content: rw.Content}
		if !strings.HasSuffix(r.Name, suffix) {
			continue
		}
		if !want[r] || have[r] {
			del = append(del, rw)
			continue
		}
		have[r] = true
	}
	for _, r := range desired {
		if !have[r] {
			create = append(create, r)
		}
	}
	return create, del
}

type Syncer struct {
	Tailscale *tailscale.Client
	NextDNS   *nextdns.Client
	Domain    string
	// DryRun logs planned changes without applying them. @default false
	DryRun bool
}

func (s *Syncer) Run(ctx context.Context) error {
	devices, err := s.Tailscale.Devices(ctx)
	if err != nil {
		return fmt.Errorf("list tailscale devices: %w", err)
	}
	current, err := s.NextDNS.Rewrites(ctx)
	if err != nil {
		return fmt.Errorf("list nextdns rewrites: %w", err)
	}
	desired := Desired(devices, s.Domain)
	create, del := Plan(desired, current, s.Domain)
	log.Printf("devices=%d rewrites=%d create=%d delete=%d dry_run=%t", len(devices), len(current), len(create), len(del), s.DryRun)

	if s.DryRun {
		pending := map[Record]bool{}
		for _, r := range create {
			pending[r] = true
		}
		for _, r := range desired {
			if !pending[r] {
				log.Printf("ok     %s -> %s", r.Name, r.Content)
			}
		}
	}

	// Creating before deleting means an address change briefly returns both IPs instead of none.
	var errs []error
	for _, r := range create {
		log.Printf("create %s -> %s", r.Name, r.Content)
		if !s.DryRun {
			if err := s.NextDNS.Create(ctx, r.Name, r.Content); err != nil {
				errs = append(errs, fmt.Errorf("create %s -> %s: %w", r.Name, r.Content, err))
			}
		}
	}
	for _, rw := range del {
		log.Printf("delete %s -> %s", rw.Name, rw.Content)
		if !s.DryRun {
			if err := s.NextDNS.Delete(ctx, rw.ID); err != nil {
				errs = append(errs, fmt.Errorf("delete %s -> %s: %w", rw.Name, rw.Content, err))
			}
		}
	}
	return errors.Join(errs...)
}
