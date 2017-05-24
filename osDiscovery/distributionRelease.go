package osDiscovery

import (
	"fmt"
	"regexp"
	"strings"
)

var distributionSpecificFiles = []string{
	"/etc/issue",
	"/etc/centos-release",
	"/etc/redhat-release",
	"/etc/SuSE-release",
	"/etc/system-release",
	"/etc/system-release-cpe",
}

var distributionSpecificFilePrefixToDistroId = map[string]string{
	"Debian":        "debian",
	"Ubuntu":        "ubuntu",
	"CentOS":        "centos",
	"cpe:/o:centos": "centos",
	"Red Hat":       "rhel",
	"cpe:/o:redhat": "rhel",
	"Amazon":        "amzn",
	"cpe:/o:amazon": "amzn",
	"openSUSE":      "opensuse",
	"SUSE":          "sles",
}

var lsbDistributionNameToIdDistributionName = map[string]string{
	"Debian":                 "debian",
	"Ubuntu":                 "ubuntu",
	"RedHatEnterpriseServer": "rhel",
	"CentOS":                 "centos",
	"openSUSE project":       "opensuse",
	"SUSE":                   "sles",
	"AmazonAMI":              "amzn",
}

// returns error if value of distribution or release is an empty string
func checkDistributionReleaseValue(distribution, release string) (string, string, error) {
	if distribution == "" {
		return "", "", ErrUnknownDistribution
	}

	if release == "" {
		return "", "", ErrUnknownRelease
	}

	return distribution, release, nil
}

// parseLsbReleaseContent returns value for 'Distributor ID' and 'Release' fields
// from content returns by lsb_release command.
func parseLsbReleaseContent(content string) (string, string, error) {
	distribution, release := getValues("Distributor ID:", "Release:", content)

	if d, ok := lsbDistributionNameToIdDistributionName[distribution]; ok {
		distribution = d
	}

	if distribution == "debian" || distribution == "centos" {
		release = strings.Split(release, ".")[0]
	}

	return checkDistributionReleaseValue(distribution, release)
}

// getValues returns values for provided fields from content.
func getValues(regexpField1, regexpField2, content string) (string, string) {
	var value1, value2 string

	matches := regexp.MustCompile(fmt.Sprintf("(%s|%s)(.*)", regexpField1, regexpField2)).FindAllStringSubmatch(content, -1)
	for _, v := range matches {
		switch {
		case v[1] == regexpField1:
			value1 = strings.Trim(v[2], " \t\" ")
		case v[1] == regexpField2:
			value2 = strings.Trim(v[2], " \t\" ")
		}
	}

	return value1, value2
}

// parseSpecificFileContent returns distribution and release if file content prefix
// matches distribution specific file prefix. Otherwise an error is returned
func parseSpecificFileContent(content string) (string, string, error) {
	for k, v := range distributionSpecificFilePrefixToDistroId {
		if strings.HasPrefix(content, k) {
			if distribution, release := getDistributionReleaseFromSpecificFileContent(v, content); release != "" {
				return distribution, release, nil
			}
		}
	}

	return "", "", ErrUnknownDistribution
}

func getDistributionReleaseFromSpecificFileContent(distroId, content string) (string, string) {
	switch distroId {
	case "ubuntu":
		return "ubuntu", getReleaseFromSpecificFileContentDefault("\\d+\\.\\d+", content)
	case "debian":
		return "debian", getReleaseFromSpecificFileContentDefault("\\d+", content)
	case "centos":
		return "centos", getReleaseFromSpecificFileContentDefault("\\d+", content)
	case "rhel":
		return "rhel", getReleaseFromSpecificFileContentDefault("\\d+\\.\\d+", content)
	case "amzn":
		return "amzn", getReleaseFromSpecificFileContentDefault("\\d+\\.\\d+", content)
	case "opensuse":
		return "opensuse", getReleaseFromSpecificFileContentDefault("\\d+\\.\\d+", content)
	case "sles":
		return "sles", getReleaseFromSpecificFileContentSles(content)
	default:
		return "", ""
	}
}

/* helper functions to get release from specific file content */

func getReleaseFromSpecificFileContentDefault(regexpRelease, content string) string {
	return regexp.MustCompile(regexpRelease).FindString(content)
}

func getReleaseFromSpecificFileContentSles(content string) string {
	version, patch := getValues("VERSION =", "PATCHLEVEL =", content)
	if version != "unknown" && patch != "unknown" && patch != "0" {
		return fmt.Sprintf("%s.%s", version, patch)
	}

	return version
}
