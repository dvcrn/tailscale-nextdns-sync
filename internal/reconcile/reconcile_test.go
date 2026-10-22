package reconcile

import (
	"slices"
	"testing"

	"github.com/dvcrn/tailscale-nextdns-sync/internal/nextdns"
	"github.com/dvcrn/tailscale-nextdns-sync/internal/tailscale"
)

func TestDesired(t *testing.T) {
	devices := []tailscale.Device{
		{Name: "comet.example.ts.net", Addresses: []string{"100.64.0.1", "fd7a:115c:a1e0::1"}},
		{Name: "Nova.example.ts.net", Addresses: []string{"100.64.0.2"}},
		{Name: "shared.other.ts.net", Addresses: []string{"100.64.0.3"}, IsExternal: true},
	}
	got := Desired(devices, "ts.d.pn")
	want := []Record{
		{"comet.ts.d.pn", "100.64.0.1"},
		{"comet.ts.d.pn", "fd7a:115c:a1e0::1"},
		{"nova.ts.d.pn", "100.64.0.2"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPlan(t *testing.T) {
	desired := []Record{
		{"comet.ts.d.pn", "100.64.0.1"},
		{"nova.ts.d.pn", "100.64.0.2"},
	}
	current := []nextdns.Rewrite{
		{ID: "keep", Name: "comet.ts.d.pn", Content: "100.64.0.1"},
		{ID: "dup", Name: "comet.ts.d.pn", Content: "100.64.0.1"},
		{ID: "stale", Name: "nova.ts.d.pn", Content: "100.64.0.9"},
		{ID: "gone", Name: "old.ts.d.pn", Content: "100.64.0.5"},
		{ID: "foreign", Name: "router.lan", Content: "192.168.1.1"},
		{ID: "apex", Name: "ts.d.pn", Content: "1.2.3.4"},
	}
	create, del := Plan(desired, current, "ts.d.pn")

	if want := []Record{{"nova.ts.d.pn", "100.64.0.2"}}; !slices.Equal(create, want) {
		t.Fatalf("create = %v, want %v", create, want)
	}
	var ids []string
	for _, rw := range del {
		ids = append(ids, rw.ID)
	}
	if want := []string{"dup", "stale", "gone"}; !slices.Equal(ids, want) {
		t.Fatalf("delete = %v, want %v", ids, want)
	}
}
