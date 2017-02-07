// +build debian ubuntu

package main

import (
	"regexp"
	"strings"
)

var (
	dpkgStatusPath  = "/var/lib/dpkg/status"
	sourcesListPath = "/etc/apt/sources.list"
)

func GetPackages() ([]*Package, error) {
	// read status of all installed packages
	data, err := readFile(dpkgStatusPath)
	if err != nil {
		log.Warning("Unable to open %s, aborting.", dpkgStatusPath)
		return nil, err
	}

	result := []*Package{}
	officialMap := map[string]*Package{}

	// split content by double newline to get separate packages
	packages := strings.Split(string(data), "\n\n")

	// iterate through packages and collect the data
	for _, pkg := range packages {
		// pattern for all key=value would be "([a-zA-Z-]+): ?(.*)"
		pattern := "(Status|Package|Version|Architecture|Source|Maintainer): (.*)"
		m := regexp.MustCompile(pattern).FindAllStringSubmatch(pkg, -1)

		// convertc matches to a map
		matches := make(map[string]string)
		for _, v := range m {
			matches[v[1]] = v[2]
		}

		if matches["Package"] != "" && matches["Status"] == "install ok installed" {
			pkg := &Package{
				Name:         matches["Package"],
				Version:      matches["Version"],
				Architecture: matches["Architecture"],
				Maintainer:   matches["Maintainer"],
				Source:       extractPackageNameFromSource(matches["Source"]),
			}
			result = append(result, pkg)
			officialMap[matches["Package"]] = pkg
		}
	}

	// check for official packages
	if err := setOfficial(officialMap); err != nil {
		return nil, err
	}

	return result, nil
}

func setOfficial(officialMap map[string]*Package) error {
	// get "official" repositories
	officialRepos, err := getRepositoriesFromSourcesList()
	if err != nil {
		return err
	}

	// collect package repositories
	policy, err := getAptCachePolicy(officialMapToList(officialMap))
	if err != nil {
		return err
	}

	// iterate over packages
	for pkgName, pkg := range officialMap {
		// 1. check the source
		if sources, ok := policy[pkgName]; ok {
			if isPackageSourceFromOfficialRepositories(sources, officialRepos) {
				pkg.Official = true
				continue
			}
		}

		// 2. check maintainer
		if aptMaintainerRe.MatchString(pkg.Maintainer) {
			pkg.Official = true
			continue
		}

		// 3. is it an official patch patch
		if aptPatchRe.MatchString(pkg.Version) {
			pkg.Official = true
			continue
		}
	}

	return nil
}

// extract keys from the map of all packages
func officialMapToList(officialMap map[string]*Package) []string {
	list := []string{}
	for pkgName, _ := range officialMap {
		list = append(list, pkgName)
	}
	return list
}

var sourceRE = regexp.MustCompile("^([\\w-\\.]+)")

// Ensures source only contains package name (ex: "shadow" and not
// "shadow (1.2.3-1)")
func extractPackageNameFromSource(source string) string {
	return strings.Split(source, " ")[0]
}

// Collects all repositories from /etc/apt/sources.list that will be treated as
// official repositories.
func getRepositoriesFromSourcesList() ([]string, error) {
	// Collect all "official" repositories from /etc/apt/sources.list
	data, err := readFile(sourcesListPath)
	if err != nil {
		log.Warningf("Unable to open %s aborting", sourcesListPath)
		return nil, err
	}

	// split the file by lines
	lines := strings.Split(string(data), "\n")

	// iterate lines and collect repository URLs
	re := regexp.MustCompile("^\\s*(deb|deb-src) ([^ ]+) ")
	repositories := map[string]struct{}{}
	for _, line := range lines {
		matches := re.FindAllStringSubmatch(line, -1)

		// skip lines with no results
		if len(matches) == 0 {
			continue
		}

		// add repository to a map, also strips trailing "/" for easier comparison
		repositories[strings.Trim(matches[0][2], "/")] = struct{}{}
	}

	// get keys from repositories map
	keys := []string{}
	for key, _ := range repositories {
		keys = append(keys, key)
	}
	return keys, nil
}

// Collects packages and their sources from `apt-cache policy`
func getAptCachePolicy(packages []string) (map[string][]string, error) {
	// Call `apt-cache policy pkg1 pkg2 ...` to collect possible package sources.
	args := append([]string{"policy"}, packages...)
	output, err := execCommand("apt-cache", args...)
	if err != nil {
		return nil, err
	}

	// convert output of `apt-cache policy` to a map[pkgName][]string
	lines := strings.Split(string(output), "\n")
	reIsPackageName := regexp.MustCompile("^([^ ]+):$")
	reVersionLine := regexp.MustCompile("^[ ]{5}\\w")
	reActiveVersionLine := regexp.MustCompile("^ \\*\\*\\*")
	reSourceLine := regexp.MustCompile("^[\\s]{8}\\d{3}")

	var pkgName string
	var installedVersion bool
	policy := map[string][]string{}

	for _, line := range lines {
		// check if a new package started
		// "pkg-name:"
		pkgNameMatches := reIsPackageName.FindAllStringSubmatch(line, -1)
		if len(pkgNameMatches) > 0 {
			pkgName = pkgNameMatches[0][1]
			policy[pkgName] = []string{}
			installedVersion = false
			continue
		}

		// are we inside a installed version block
		if installedVersion {
			// did a new version start
			if reVersionLine.MatchString(line) {
				installedVersion = false
				continue
			}

			// we should have a pkg source line h
			if reSourceLine.MatchString(line) {
				lineItems := strings.Split(strings.Trim(line, " "), " ")
				// skip local sources (/var/lib/dpkg/status)
				if lineItems[1] == "/var/lib/dpkg/status" {
					continue
				}

				// to make things easier to compare we strip tailing "/" from the source
				policy[pkgName] = append(policy[pkgName], strings.Trim(lineItems[1], "/"))
			}
			continue
		}

		if reActiveVersionLine.MatchString(line) {
			installedVersion = true
		}
	}

	return policy, nil
}

// checks if any of the strings in pkgSources is present in official repositories
func isPackageSourceFromOfficialRepositories(pkgSources, officialRepos []string) bool {
	// stop when one of the lists empty
	if len(pkgSources) == 0 || len(officialRepos) == 0 {
		return false
	}

	// convert officialRepos to a map to ease comparison
	officialMap := map[string]struct{}{}
	for _, repo := range officialRepos {
		officialMap[repo] = struct{}{}
	}

	// find the first matching item
	for _, source := range pkgSources {
		if _, ok := officialMap[source]; ok {
			return true
		}
	}

	return false
}
