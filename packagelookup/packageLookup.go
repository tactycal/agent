// Package packagelookup provides function to get list of installed packages.
// Supported operating systems:
//
//  - Ubuntu
//  - Debian
//  - Red Hat Enterprise Linux
//  - CentOS
//  - Amazon Linux AMI
//  - openSUSE
//  - SUSE Linux Enterprise Server
package packagelookup

import "errors"

// Constants show supported operating systems and can be used as osID argument
// in Get function.
const (
	// Ubuntu
	UBUNTU = "ubuntu"
	// Debian
	DEBIAN = "debian"
	// Red Hat Enterprise Linux
	RHEL = "rhel"
	// CentOS
	CENTOS = "centos"
	// Amazon Linux AMI
	AMZN = "amzn"
	// openSUSE
	OPENSUSE = "opensuse"
	// SUSE Linux Enterprise Server
	SLES = "sles"
)

// List of possible errors
var (
	ErrDistributionNotSupported = errors.New("Distribution is not supported")
)

// Package represent installed package.
type Package struct {
	// Package name.
	Name string `json:"name"`
	// Package version.
	Version string `json:"version"`
	// Name of the source package.
	Source string `json:"source"`
	// Package architecture.
	Architecture string `json:"architecture"`
	// Set to true if package is installed from operating system distributor
	// repositories or vendor of package is operating system distributor.
	// Default value is false.
	Official   bool `json:"official"`
	maintainer string
}

// Get returns installed packages for a given operating system. osID argument must
// be valid operating system distributor id provided in /etc/os-release file.
func Get(osID string) ([]*Package, error) {
	switch osID {
	case UBUNTU:
		return getApt(aptMaintainerUbuntu, aptPatchUbuntu)
	case DEBIAN:
		return getApt(aptMaintainerDebian, aptPatchDebian)
	case RHEL:
		return getRpm(rmpVendorRhel)
	case CENTOS:
		return getRpm(rpmVendorCentos)
	case AMZN:
		return getRpm(rpmVendorAmzn)
	case OPENSUSE:
		return getRpm(rpmVendorOpensuse)
	case SLES:
		return getRpm(rpmVendorSles)
	default:
		return nil, ErrDistributionNotSupported
	}
}
