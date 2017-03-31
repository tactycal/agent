// +build opensuse sles amzn

package main

import (
	"fmt"
	"regexp"
	"strings"
)

type mappedPackages map[string]*Package

func GetPackages() ([]*Package, error) {
	return readPackages()
}

// Fetches packages with `rpm -qa -queryformat`
func readPackages() ([]*Package, error) {
	results := []*Package{}

	b, err := execCommand(`rpm`, `-qa`, `--queryformat`, `Name: %{NAME}\nArch: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`)
	if err != nil {
		return results, err
	}

	// split response by packages
	packages := strings.Split(string(b), "\n\n")

	// loop through packages
	reKV := regexp.MustCompile("(Name|Arch|Version|Release|Vendor|Source|Epoch): (.*)")
	for _, pkg := range packages {
		m := reKV.FindAllStringSubmatch(pkg, -1)

		// put results into a key value map
		matches := make(map[string]string)
		for _, v := range m {
			matches[v[1]] = v[2]
		}

		// skip empty lines
		if matches["Name"] == "" {
			continue
		}

		results = append(results, buildPackage(matches))
	}

	return results, nil
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
