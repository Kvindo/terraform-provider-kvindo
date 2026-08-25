package provider

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// kcProfile mirrors the subset of KvindoCloud.CLI's Profile struct (config.go) the
// provider needs to resolve an endpoint/token from a profile `kc login`/`kc switch`
// already wrote. Kept as a local copy since the CLI and this provider are separate Go
// modules with no shared package to import from.
type kcProfile struct {
	Server string `yaml:"server"`
	Token  string `yaml:"token"`
}

// loadKcProfile reads ~/.kc/config/<name>.yaml. It returns a zero-value kcProfile (no
// error) when the file is missing, unreadable, or malformed — a bad/absent cli_profile
// falls through to the endpoint/token env-var and default resolution in Configure
// rather than hard-failing the whole provider.
func loadKcProfile(name string) kcProfile {
	home, err := os.UserHomeDir()
	if err != nil {
		return kcProfile{}
	}
	data, err := os.ReadFile(filepath.Join(home, ".kc", "config", name+".yaml"))
	if err != nil {
		return kcProfile{}
	}
	var p kcProfile
	_ = yaml.Unmarshal(data, &p)
	return p
}
