// +build ubuntu

package main

import "regexp"

const (
	DISTRIBUTION = "ubuntu"
)

var aptMaintainerRe = regexp.MustCompile("(ubuntu.com|canonical.com|debian.org)")
var aptPatchRe = regexp.MustCompile("-[\\d\\.]+ubuntu[\\d\\.~]+$")
