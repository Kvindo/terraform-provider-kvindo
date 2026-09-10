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

// loadKcProfile reads ~/.kc/config/<name>.yaml. name is sanitized with filepath.Base
// first, since it can come from HCL config or an env var and is joined directly into a
// filesystem path. It returns a zero-value kcProfile and found=false when the file is
// missing or unreadable — a bad/absent cli_profile falls through to the endpoint/token
// env-var and default resolution in Configure rather than hard-failing the whole
// provider. found=true with a zero-value kcProfile means the file was read but didn't
// parse as valid YAML.
func loadKcProfile(name string) (profile kcProfile, found bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return kcProfile{}, false
	}
	data, err := os.ReadFile(filepath.Join(home, ".kc", "config", filepath.Base(name)+".yaml"))
	if err != nil {
		return kcProfile{}, false
	}
	var p kcProfile
	_ = yaml.Unmarshal(data, &p)
	return p, true
}
