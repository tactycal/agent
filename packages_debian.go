// +build debian

package main

import "regexp"

const (
	DISTRIBUTION = "debian"
)

var aptMaintainerRe = regexp.MustCompile("debian.org")
var aptPatchRe = regexp.MustCompile("\\+deb\\d+u\\d+$")
