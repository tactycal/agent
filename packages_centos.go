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
		Official:     isOfficial(matches["From repo"]),
	}
}

func isOfficial(fromRepo string) bool {
	return fromRepo == "CentOS" || fromRepo == "Updates" || fromRepo == "base"
}
