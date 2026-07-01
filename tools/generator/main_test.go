package main

import "testing"

// camelToSnake must keep IP-version suffixes as one segment (public_ipv4, not public_ip_v4) so the
// generated Terraform schema keys match the platform-wide convention. Regression guard for the
// ipVersionAcronyms normalization.
func TestCamelToSnakeIpVersion(t *testing.T) {
	cases := map[string]string{
		"publicIpV4":       "public_ipv4",
		"privateIpV6":      "private_ipv6",
		"natPublicIpV4":    "nat_public_ipv4",
		"sshIpV4":          "ssh_ipv4",
		"ipV4Cidrs":        "ipv4_cidrs",
		"allowedIpV4Cidrs": "allowed_ipv4_cidrs",
		"assignPublicIpV4": "assign_public_ipv4",
		"createPublicIpv4": "create_public_ipv4", // already-correct form is left untouched
		"volumeSizeGiB":    "volume_size_gib",     // existing GiB normalization still holds
	}
	for in, want := range cases {
		if got := camelToSnake(in); got != want {
			t.Errorf("camelToSnake(%q) = %q, want %q", in, got, want)
		}
	}
}
