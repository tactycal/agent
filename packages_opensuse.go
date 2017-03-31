// +build opensuse

package main

const (
	DISTRIBUTION = "opensuse"
)

func buildPackage(matches map[string]string) *Package {
	return &Package{
		Name:         matches["Name"],
		Version:      buildVersion(matches),
		Architecture: matches["Arch"],
		Official:     matches["Vendor"] == "openSUSE",
		Source:       getSourceName(matches),
	}
}
