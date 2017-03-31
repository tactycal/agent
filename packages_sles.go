// +build sles

package main

const (
	DISTRIBUTION = "sles"
)

func buildPackage(matches map[string]string) *Package {
	return &Package{
		Name:         matches["Name"],
		Version:      buildVersion(matches),
		Architecture: matches["Arch"],
		Official:     matches["Vendor"] == "SUSE LLC <https://www.suse.com/>",
		Source:       getSourceName(matches),
	}
}
