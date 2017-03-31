// +build sles

package main

import (
	"fmt"
	"regexp"
)

var (
	// SuSE-release file
	releaseSuseRE = regexp.MustCompile("(VERSION|PATCHLEVEL) = (\\d+)")

	// os-release file or lsb_release
	releaseOsRE = regexp.MustCompile("(\\d+(\\.\\d+)?)")
)

func getHostRelease() string {
	hostRelease := readHostRelease()

	// SuSE-release
	var version, patch string
	for _, v := range releaseSuseRE.FindAllStringSubmatch(hostRelease, 2) {
		switch v[1] {
		case "VERSION":
			version = v[2]
		case "PATCHLEVEL":
			patch = v[2]
		}
	}

	if version != "" && patch != "" {
		if patch == "0" {
			return version
		}

		return fmt.Sprintf("%s.%s", version, patch)
	}

	// os-release
	matches := releaseOsRE.FindStringSubmatch(hostRelease)
	if len(matches) > 0 {
		return matches[1]
	}

	return "unknown"
}
