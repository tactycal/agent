// Package osDiscovery implements functions to get basic operating system
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
package osDiscovery

import "errors"

// OsInfo contains information about the operating system.
type OsInfo struct {
	Distribution string `json:"distribution"`
	Release      string `json:"release"`
	Architecture string `json:"architecture"`
	Kernel       string `json:"kernel"`
	Fqdn         string `json:"fqdn"`
}

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
	if out, err = readFile("/etc/os-release"); err == nil {
		distribution, release := getValues("ID=", "VERSION_ID=", string(out))
		return checkDistributionReleaseValue(distribution, release)
	}

	// fallback to LSB
	if out, err = execCommand("lsb_release", "-ir"); err == nil {
		return parseLsbReleaseContent(string(out))
	}

	// distro specific
	for _, f := range distributionSpecificFiles {
		if out, err = readFile(f); err == nil {
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
	if out, err := execCommand("uname", "-m"); err == nil {
		return string(out), nil
	}
	return "", ErrUnknownArchitecture
}

// GetKernel returns a kernel release.
func GetKernel() (string, error) {
	if out, err := execCommand("uname", "-r"); err == nil {
		return string(out), nil
	}
	return "", ErrUnknownKernel
}

// GetFqdn returns the fully qualified domain name (FQDN).
func GetFqdn() (string, error) {
	if out, err := execCommand("hostname", "-f"); err == nil {
		return string(out), nil
	}

	if out, err := readFile("/etc/hostname"); err == nil {
		return string(out), nil
	}

	return "", ErrUnknownFqdn
}
