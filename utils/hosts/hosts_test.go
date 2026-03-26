package hosts

import (
	"strings"
	"testing"
)

const sampleHosts = `##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting.  Do not alter this content.
##
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost
# Content below this line is managed by vpd hosts.
# DO NOT EDIT.
# profile.on myproject
127.0.0.1	myapp.local
127.0.0.1	api.local
# end
# profile.off staging
# 10.0.0.1	staging.local
# 10.0.0.1	staging-api.local
# end
`

func TestParseContent(t *testing.T) {
	hf, err := ParseContent(sampleHosts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hf.Preamble) != 9 {
		t.Errorf("expected 9 preamble lines, got %d", len(hf.Preamble))
	}

	if len(hf.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(hf.Profiles))
	}

	p1 := hf.Profiles[0]
	if p1.Name != "myproject" || !p1.Enabled {
		t.Errorf("expected profile 'myproject' enabled, got %q enabled=%v", p1.Name, p1.Enabled)
	}
	if len(p1.Entries) != 2 {
		t.Errorf("expected 2 entries in myproject, got %d", len(p1.Entries))
	}

	p2 := hf.Profiles[1]
	if p2.Name != "staging" || p2.Enabled {
		t.Errorf("expected profile 'staging' disabled, got %q enabled=%v", p2.Name, p2.Enabled)
	}
	if len(p2.Entries) != 2 {
		t.Errorf("expected 2 entries in staging, got %d", len(p2.Entries))
	}
}

func TestRenderRoundtrip(t *testing.T) {
	hf, err := ParseContent(sampleHosts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rendered := hf.Render()

	hf2, err := ParseContent(rendered)
	if err != nil {
		t.Fatalf("unexpected error on re-parse: %v", err)
	}

	if len(hf2.Profiles) != len(hf.Profiles) {
		t.Errorf("profile count mismatch: %d vs %d", len(hf.Profiles), len(hf2.Profiles))
	}

	for i, p := range hf.Profiles {
		p2 := hf2.Profiles[i]
		if p.Name != p2.Name || p.Enabled != p2.Enabled || len(p.Entries) != len(p2.Entries) {
			t.Errorf("profile %d mismatch", i)
		}
	}
}

func TestAddEntries(t *testing.T) {
	hf, _ := ParseContent(sampleHosts)
	hf.AddEntries("newprofile", []Entry{{IP: "192.168.1.1", Hostname: "new.local"}})

	if len(hf.Profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(hf.Profiles))
	}
	p := hf.GetProfile("newprofile")
	if p == nil || len(p.Entries) != 1 {
		t.Fatal("new profile not created correctly")
	}

	// Add to existing
	hf.AddEntries("myproject", []Entry{{IP: "127.0.0.1", Hostname: "extra.local"}})
	p = hf.GetProfile("myproject")
	if len(p.Entries) != 3 {
		t.Errorf("expected 3 entries after add, got %d", len(p.Entries))
	}
}

func TestRemoveProfile(t *testing.T) {
	hf, _ := ParseContent(sampleHosts)
	if !hf.RemoveProfile("myproject") {
		t.Error("expected RemoveProfile to return true")
	}
	if len(hf.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(hf.Profiles))
	}
	if !hf.RemoveProfile("staging") {
		t.Error("expected RemoveProfile to return true")
	}
	if len(hf.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(hf.Profiles))
	}
	// No managed section in render
	rendered := hf.Render()
	if strings.Contains(rendered, "managed by vpd") {
		t.Error("expected no managed section when no profiles")
	}
}

func TestRemoveEntry(t *testing.T) {
	hf, _ := ParseContent(sampleHosts)
	if !hf.RemoveEntry("myproject", "api.local") {
		t.Error("expected RemoveEntry to return true")
	}
	p := hf.GetProfile("myproject")
	if len(p.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(p.Entries))
	}
}

func TestEnableDisable(t *testing.T) {
	hf, _ := ParseContent(sampleHosts)

	if !hf.DisableProfile("myproject") {
		t.Error("expected DisableProfile to return true")
	}
	p := hf.GetProfile("myproject")
	if p.Enabled {
		t.Error("expected profile to be disabled")
	}

	rendered := hf.Render()
	if !strings.Contains(rendered, "# profile.off myproject") {
		t.Error("expected disabled marker in rendered output")
	}
	if !strings.Contains(rendered, "# 127.0.0.1\tmyapp.local") {
		t.Error("expected commented entry in rendered output")
	}

	if !hf.EnableProfile("staging") {
		t.Error("expected EnableProfile to return true")
	}
	p2 := hf.GetProfile("staging")
	if !p2.Enabled {
		t.Error("expected staging to be enabled")
	}
}

func TestEmptyHostsFile(t *testing.T) {
	hf, err := ParseContent("127.0.0.1\tlocalhost\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hf.Preamble) != 1 {
		t.Errorf("expected 1 preamble line, got %d", len(hf.Preamble))
	}
	if len(hf.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(hf.Profiles))
	}
}
