// +build rhel

package main

const (
	DISTRIBUTION = "rhel"
)

func buildPackage(matches map[string]string) *Package {
	return &Package{
		Name:         matches["Name"],
		Version:      buildVersion(matches),
		Architecture: matches["Arch"],
		Official:     matches["Vendor"] == "Red Hat, Inc.",
		Source:       getSourceName(matches),
	}
}
