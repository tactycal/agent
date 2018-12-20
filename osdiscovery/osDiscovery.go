// Package osdiscovery implements functions to get basic operating system
// identification data for Linux operating system. The following informations
// are provided:
//
//   - Distribution name
//   - Release version
//   - Architecture
//   - Fully qualified domain name
//   - Kernel release
//
// If any of the called function could not retrieve the information the error
// will be returned.
package osdiscovery

import (
	"errors"
	"strings"

	"github.com/tactycal/agent/stubutils"
)

// OsInfo contains information about the operating system.
type OsInfo struct {
	Distribution string `json:"distribution"`
	Release      string `json:"release"`
	Architecture string `json:"architecture"`
	Kernel       string `json:"kernel"`
	Fqdn         string `json:"fqdn"`
}

// List of possible errors
var (
	ErrUnknownDistribution = errors.New("Unknown distribution")
	ErrUnknownRelease      = errors.New("Unknown release")
	ErrUnknownArchitecture = errors.New("Unknown architecture")
	ErrUnknownFqdn         = errors.New("Unknown fqdn")
	ErrUnknownKernel       = errors.New("Unknown kernel")
)

// Get fetch all the information that OsInfo provides. If any of the
// information could not be retrieved the corresponding error is returned.
func Get() (*OsInfo, error) {
	distribution, release, err := GetDistributionRelease()
	if err != nil {
		return nil, err
	}

	architecture, err := GetArchitecture()
	if err != nil {
		return nil, err
	}

	kernel, err := GetKernel()
	if err != nil {
		return nil, err
	}

	fqdn, err := GetFqdn()
	if err != nil {
		return nil, err
	}

	osInfo := &OsInfo{
		Distribution: distribution,
		Release:      release,
		Architecture: architecture,
		Kernel:       kernel,
		Fqdn:         fqdn,
	}

	return osInfo, nil
}

// GetDistributionRelease returns distribution name and release version of the
// operating system.
func GetDistributionRelease() (string, string, error) {
	var out []byte
	var err error

	// os-release
	if out, err = stubutils.ReadFile("/etc/os-release"); err == nil {
		distribution, release := getValues("ID=", "VERSION_ID=", string(out))
		return checkDistributionReleaseValue(distribution, release)
	}

	// fallback to LSB
	if out, err = stubutils.ExecCommand("lsb_release", "-ir"); err == nil {
		return parseLsbReleaseContent(string(out))
	}

	// distro specific
	for _, f := range distributionSpecificFiles {
		if out, err = stubutils.ReadFile(f); err == nil {
			distribution, release, err := parseSpecificFileContent(string(out))
			if err == nil {
				return distribution, release, nil
			}
		}
	}

	return "", "", ErrUnknownDistribution
}

// GetArchitecture returns a machine hardware name.
func GetArchitecture() (string, error) {
	if out, err := stubutils.ExecCommand("uname", "-m"); err == nil {
		return strings.TrimSuffix(string(out), "\n"), nil
	}
	return "", ErrUnknownArchitecture
}

// GetKernel returns a kernel release.
func GetKernel() (string, error) {
	if out, err := stubutils.ExecCommand("uname", "-r"); err == nil {
		return strings.TrimSuffix(string(out), "\n"), nil
	}
	return "", ErrUnknownKernel
}

// GetFqdn returns the fully qualified domain name (FQDN).
func GetFqdn() (string, error) {
	if out, err := stubutils.ExecCommand("hostname", "-f"); err == nil {
		return strings.TrimSuffix(string(out), "\n"), nil
	}

	if out, err := stubutils.ReadFile("/etc/hostname"); err == nil {
		return strings.TrimSuffix(string(out), "\n"), nil
	}

	return "", ErrUnknownFqdn
}
