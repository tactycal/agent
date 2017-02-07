// +build ubuntu debian

package main

import "regexp"

var releaseRE = regexp.MustCompile("(\\d+(\\.\\d+)?)")

func getHostRelease() string {
	matches := releaseRE.FindStringSubmatch(readHostRelease())

	if len(matches) > 0 {
		return matches[1]
	}

	return "unknown"
}
