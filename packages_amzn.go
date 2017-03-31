// +build amzn

package main

const (
	DISTRIBUTION = "amzn"
)

func buildPackage(matches map[string]string) *Package {
	return &Package{
		Name:         matches["Name"],
		Version:      buildVersion(matches),
		Architecture: matches["Arch"],
		Official:     matches["Vendor"] == "Amazon.com",
		Source:       getSourceName(matches),
	}
}
