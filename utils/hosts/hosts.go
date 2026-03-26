package hosts

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	profileOnPrefix  = "# profile.on "
	profileOffPrefix = "# profile.off "
	profileEnd       = "# end"
	ManagedHeader    = "# Content below this line is managed by vpd hosts."
)

// Entry represents a single host entry.
type Entry struct {
	IP       string
	Hostname string
	Aliases  []string
}

func (e Entry) String() string {
	parts := []string{e.IP, e.Hostname}
	parts = append(parts, e.Aliases...)
	return strings.Join(parts, "\t")
}

// Profile represents a named group of host entries.
type Profile struct {
	Name    string
	Enabled bool
	Entries []Entry
}

// HostsFile represents the parsed contents of an /etc/hosts file.
type HostsFile struct {
	// Lines before the managed section (preserved as-is).
	Preamble []string
	Profiles []Profile
}

// DefaultPath returns the platform-specific hosts file path.
func DefaultPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\Drivers\etc\hosts`
	}
	return "/etc/hosts"
}

// Parse reads and parses a hosts file.
func Parse(path string) (*HostsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading hosts file: %w", err)
	}
	return ParseContent(string(data))
}

// ParseContent parses the content of a hosts file.
func ParseContent(content string) (*HostsFile, error) {
	hf := &HostsFile{}
	scanner := bufio.NewScanner(strings.NewReader(content))

	inManaged := false
	var currentProfile *Profile

	for scanner.Scan() {
		line := scanner.Text()

		if !inManaged {
			if strings.TrimSpace(line) == ManagedHeader {
				inManaged = true
				continue
			}
			hf.Preamble = append(hf.Preamble, line)
			continue
		}

		trimmed := strings.TrimSpace(line)

		// Skip the "DO NOT EDIT" line if present
		if trimmed == "# DO NOT EDIT." {
			continue
		}

		if strings.HasPrefix(trimmed, profileOnPrefix) {
			name := strings.TrimPrefix(trimmed, profileOnPrefix)
			currentProfile = &Profile{Name: name, Enabled: true}
			continue
		}
		if strings.HasPrefix(trimmed, profileOffPrefix) {
			name := strings.TrimPrefix(trimmed, profileOffPrefix)
			currentProfile = &Profile{Name: name, Enabled: false}
			continue
		}
		if trimmed == profileEnd {
			if currentProfile != nil {
				hf.Profiles = append(hf.Profiles, *currentProfile)
				currentProfile = nil
			}
			continue
		}

		if currentProfile != nil {
			entry := parseLine(line, !currentProfile.Enabled)
			if entry != nil {
				currentProfile.Entries = append(currentProfile.Entries, *entry)
			}
		}
	}

	// Handle unclosed profile
	if currentProfile != nil {
		hf.Profiles = append(hf.Profiles, *currentProfile)
	}

	return hf, scanner.Err()
}

// parseLine parses a single host entry line. If commented is true,
// expects the line to be prefixed with "# ".
func parseLine(line string, commented bool) *Entry {
	s := strings.TrimSpace(line)
	if s == "" {
		return nil
	}
	if commented {
		s = strings.TrimPrefix(s, "# ")
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "#") {
		return nil // skip comment lines in enabled profiles
	}

	fields := strings.Fields(s)
	if len(fields) < 2 {
		return nil
	}

	if net.ParseIP(fields[0]) == nil {
		return nil
	}

	entry := &Entry{
		IP:       fields[0],
		Hostname: fields[1],
	}
	if len(fields) > 2 {
		entry.Aliases = fields[2:]
	}
	return entry
}

// Render produces the full hosts file content.
func (hf *HostsFile) Render() string {
	var b strings.Builder

	for _, line := range hf.Preamble {
		b.WriteString(line)
		b.WriteString("\n")
	}

	if len(hf.Profiles) == 0 {
		return b.String()
	}

	b.WriteString(ManagedHeader)
	b.WriteString("\n")
	b.WriteString("# DO NOT EDIT.\n")

	for _, p := range hf.Profiles {
		if p.Enabled {
			b.WriteString(profileOnPrefix + p.Name + "\n")
		} else {
			b.WriteString(profileOffPrefix + p.Name + "\n")
		}
		for _, e := range p.Entries {
			if !p.Enabled {
				b.WriteString("# " + e.String() + "\n")
			} else {
				b.WriteString(e.String() + "\n")
			}
		}
		b.WriteString(profileEnd + "\n")
	}

	return b.String()
}

// Write writes the hosts file content to the given path.
// If a direct write fails with permission denied, it automatically
// retries using sudo tee.
func (hf *HostsFile) Write(path string) error {
	content := hf.Render()
	err := os.WriteFile(path, []byte(content), 0644)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrPermission) {
		return err
	}
	return writeWithSudo(path, content)
}

func writeWithSudo(path, content string) error {
	cmd := exec.Command("sudo", "tee", path)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stderr = os.Stderr
	cmd.Stdout = nil // suppress tee's stdout echo
	return cmd.Run()
}

// GetProfile returns a pointer to the named profile, or nil if not found.
func (hf *HostsFile) GetProfile(name string) *Profile {
	for i := range hf.Profiles {
		if hf.Profiles[i].Name == name {
			return &hf.Profiles[i]
		}
	}
	return nil
}

// AddEntries adds entries to a profile, creating it if it doesn't exist.
func (hf *HostsFile) AddEntries(profileName string, entries []Entry) {
	p := hf.GetProfile(profileName)
	if p == nil {
		hf.Profiles = append(hf.Profiles, Profile{
			Name:    profileName,
			Enabled: true,
			Entries: entries,
		})
		return
	}
	p.Entries = append(p.Entries, entries...)
}

// RemoveProfile removes an entire profile by name.
func (hf *HostsFile) RemoveProfile(name string) bool {
	for i, p := range hf.Profiles {
		if p.Name == name {
			hf.Profiles = append(hf.Profiles[:i], hf.Profiles[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveEntry removes entries matching a hostname or alias from a profile.
// If the hostname matches the primary hostname, the entire entry is removed.
// If it matches an alias, only that alias is removed from the entry.
func (hf *HostsFile) RemoveEntry(profileName, hostname string) bool {
	p := hf.GetProfile(profileName)
	if p == nil {
		return false
	}
	for i, e := range p.Entries {
		if e.Hostname == hostname {
			p.Entries = append(p.Entries[:i], p.Entries[i+1:]...)
			return true
		}
		for j, alias := range e.Aliases {
			if alias == hostname {
				p.Entries[i].Aliases = append(e.Aliases[:j], e.Aliases[j+1:]...)
				return true
			}
		}
	}
	return false
}

// EnableProfile enables a profile.
func (hf *HostsFile) EnableProfile(name string) bool {
	p := hf.GetProfile(name)
	if p == nil {
		return false
	}
	p.Enabled = true
	return true
}

// DisableProfile disables a profile.
func (hf *HostsFile) DisableProfile(name string) bool {
	p := hf.GetProfile(name)
	if p == nil {
		return false
	}
	p.Enabled = false
	return true
}
