package packageLookup

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tactycal/agent/stubUtils"
)

const (
	rmpVendorRhel     = "Red Hat, Inc."
	rpmVendorCentos   = "CentOS"
	rpmVendorAmzn     = "Amazon.com"
	rpmVendorOpensuse = "openSUSE"
	rpmVendorSles     = "SUSE LLC <https://www.suse.com/>"
)

// returns packages for distribution using a RPM package manager
func getRpm(rpmVendor string) ([]*Package, error) {
	b, err := stubUtils.ExecCommand(`rpm`, `-qa`, `--queryformat`, `Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`)

	if err != nil {
		return nil, err
	}
	// split content by double newline to get separate packages
	packages := strings.Split(string(b), "\n\n")

	result := []*Package{}
	// pattern for all key=value would be "([a-zA-Z-]+): ?(.*)"
	reKV := regexp.MustCompile("(Name|Version|Release|Vendor|Architecture|Source|Epoch): (.*)")
	// iterate through packages and collect the data
	for _, pkg := range packages {
		m := reKV.FindAllStringSubmatch(pkg, -1)

		// convert matches to a map
		matches := make(map[string]string)
		for _, v := range m {
			matches[v[1]] = v[2]
		}

		if matches["Name"] != "" {
			p := &Package{
				Name:         matches["Name"],
				Architecture: matches["Architecture"],
				Version:      buildVersion(matches),
				Source:       getSourceName(matches),
				Official:     matches["Vendor"] == rpmVendor,
			}

			result = append(result, p)
		}
	}

	return result, nil
}

func buildVersion(matches map[string]string) string {
	version := ""

	if epoch, ok := matches["Epoch"]; ok && epoch != "(none)" {
		version = fmt.Sprintf("%s:", epoch)
	}

	return fmt.Sprintf("%s%s-%s", version, matches["Version"], matches["Release"])
}

func getSourceName(matches map[string]string) string {
	ss := strings.Split(matches["Source"], "-")
	if len(ss) < 3 {
		return "unknown"
	}

	return strings.Join(ss[:len(ss)-2], "-")
}
