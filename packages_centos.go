// +build centos

package main

const (
	DISTRIBUTION = "centos"
)

func buildPackage(matches map[string]string) *Package {
	return &Package{
		Name:         matches["Name"],
		Version:      buildVersion(matches),
		Architecture: matches["Arch"],
		Official:     matches["Vendor"] == "CentOS",
		Source:       getSourceName(matches),
	}
}
