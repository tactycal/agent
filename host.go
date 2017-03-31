package main

import (
	"regexp"
	"strings"
)

type Host struct {
	Fqdn         string   `json:"fqdn"`
	Distribution string   `json:"distribution"`
	Release      string   `json:"release"`
	Architecture string   `json:"architecture"`
	Kernel       string   `json:"kernel"`
	Labels       []string `json:"labels"`
}

func GetHostInfo() Host {
	return Host{
		Fqdn:         strings.TrimSpace(getHostFqdn()),
		Distribution: DISTRIBUTION,
		Release:      strings.TrimSpace(getHostRelease()),
		Architecture: strings.TrimSpace(getHostArchitecture()),
		Kernel:       strings.TrimSpace(getHostKernel()),
	}
}

func getHostFqdn() string {
	if out, err := execCommand("hostname", "-f"); err == nil {
		return string(out)
	}

	if out, err := readFile("/etc/hostname"); err == nil {
		return string(out)
	}

	return "unknown"
}

func readHostRelease() string {
	// try new operating system identification standard
	// http://www.freedesktop.org/software/systemd/man/os-release.html
	if content, err := readFile("/etc/os-release"); err == nil {
		matches := regexp.MustCompile("VERSION_ID=(.*)").FindStringSubmatch(string(content))
		if matches != nil {
			return strings.Trim(matches[1], "\"")
		}
	}

	// fallback to LSB
	if release, err := execCommand("lsb_release", "-r"); err == nil {
		return string(release)
	}

	// distro specific
	files := []string{"/etc/centos-release", "/etc/redhat-release", "/etc/SuSE-release"}
	for _, file := range files {
		if content, err := readFile(file); err == nil {
			return string(content)
		}
	}

	return "unknown"
}

func getHostArchitecture() string {
	if out, err := execCommand("uname", "-m"); err == nil {
		return string(out)
	}
	return "unknown"
}

func getHostKernel() string {
	if out, err := execCommand("uname", "-r"); err == nil {
		return string(out)
	}
	return "unknown"
}
