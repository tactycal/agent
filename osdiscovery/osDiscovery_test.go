package osdiscovery

import (
	"reflect"
	"testing"

	"github.com/tactycal/agent/stubutils"
)

func TestGet(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.ReadFileStub{Path: "/etc/os-release", Err: stubutils.ErrOhNo},
		&stubutils.CmdStub{Cmd: "lsb_release", Args: []string{"-ir"}, Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/issue", Output: []byte("unknown 8 \n \\l")},
		&stubutils.ReadFileStub{Path: "/etc/centos-release", Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/redhat-release", Output: []byte("unknown 8 \n \\l")},
		&stubutils.ReadFileStub{Path: "/etc/SuSE-release", Output: []byte("SUSE Linux Enterprise\nVERSION = 12\nPATCHLEVEL = 0\n# This file is d")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Output: []byte("ARCH")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Output: []byte("KERN")},
		&stubutils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Output: []byte("FQDN")},
	)
	defer s.Close()

	expected := &OsInfo{
		Distribution: "sles",
		Release:      "12",
		Architecture: "ARCH",
		Kernel:       "KERN",
		Fqdn:         "FQDN",
	}

	// 1: all good
	result, _ := Get()
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected osInfo\n%+v\ngot\n%+v\n", expected, result)
	}

	// 2: unknown distribution
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Err: stubutils.ErrOhNo},
		&stubutils.CmdStub{Cmd: "lsb_release", Args: []string{"-ir"}, Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/issue", Output: []byte("unknown 8 \n \\l")},
		&stubutils.ReadFileStub{Path: "/etc/centos-release", Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/redhat-release", Output: []byte("unknown 8 \n \\l")},
		&stubutils.ReadFileStub{Path: "/etc/SuSE-release", Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/system-release", Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/system-release-cpe", Err: stubutils.ErrOhNo},
	)

	_, err := Get()
	if err != ErrUnknownDistribution {
		t.Errorf("Expected error \"%s\"; got \"%s\"", ErrUnknownDistribution.Error(), err.Error())
	}

}

func TestGetDistributionRelease(t *testing.T) {
	s := stubutils.NewStubs(t)
	defer s.Close()

	// 1: os-release
	// 1.1: all good
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=ubuntu\nID_LIKE=debian\nVERSION_ID=\"14.04\"\nHOME_URL=\"http://www.ubuntu.com/\"")})

	distribution, release, _ := GetDistributionRelease()
	if distribution != "ubuntu" {
		t.Errorf("Expected distribution \"%s\"; got \"%s\"", "ubuntu", distribution)
	}
	if release != "14.04" {
		t.Errorf("Expected release \"%s\"; got \"%s\"", "14.04", release)
	}

	// 1.2: unknown distribution
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Output: []byte("\nID_LIKE=debian\nVERSION_ID=\"14.04\"\nHOME_URL=\"http://www.ubuntu.com/\"")})

	_, _, err := GetDistributionRelease()
	if err != ErrUnknownDistribution {
		t.Errorf("Expected error \"%s\"; got \"%s\"", ErrUnknownDistribution.Error(), err.Error())
	}

	// 1.3: unknown release
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=ubuntu\nID_LIKE=debian\nHOME_URL=\"http://www.ubuntu.com/\"")})

	_, _, err = GetDistributionRelease()
	if err != ErrUnknownRelease {
		t.Errorf("Expected error \"%s\"; got \"%s\"", ErrUnknownRelease.Error(), err.Error())
	}

	// 2: lsbFallback
	// 2.1: all good
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Err: stubutils.ErrOhNo},
		&stubutils.CmdStub{Cmd: "lsb_release", Args: []string{"-ir"}, Output: []byte("Description:    Red Hat \n Distributor ID: RedHatEnterpriseServer\nRelease:    7.3")})

	distribution, release, _ = GetDistributionRelease()
	if distribution != "rhel" {
		t.Errorf("Expected distribution \"%s\"; got \"%s\"", "rhel", distribution)
	}
	if release != "7.3" {
		t.Errorf("Expected release \"%s\"; got \"%s\"", "7.3", release)
	}

	// 2.2: unknown distribution
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Err: stubutils.ErrOhNo},
		&stubutils.CmdStub{Cmd: "lsb_release", Args: []string{"-ir"}, Output: []byte("Description:    Red Hat \n Missing ID: RedHatEnterpriseServer\nRelease:    7.3")})

	_, _, err = GetDistributionRelease()
	if err != ErrUnknownDistribution {
		t.Errorf("Expected error \"%s\"; got \"%s\"", ErrUnknownDistribution.Error(), err.Error())
	}

	// 2.3: unknown release
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Err: stubutils.ErrOhNo},
		&stubutils.CmdStub{Cmd: "lsb_release", Args: []string{"-ir"}, Output: []byte("Description:    Red Hat \n Distributor ID: RedHatEnterpriseServer\nMissing:    7.3")})

	_, _, err = GetDistributionRelease()
	if err != ErrUnknownRelease {
		t.Errorf("Expected error \"%s\"; got \"%s\"", ErrUnknownRelease.Error(), err.Error())
	}

	// 3: distribution specific
	// 3.1: all good
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Err: stubutils.ErrOhNo},
		&stubutils.CmdStub{Cmd: "lsb_release", Args: []string{"-ir"}, Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/issue", Output: []byte("Debian GNU/Linux 8 \n \\l")})

	distribution, release, _ = GetDistributionRelease()
	if distribution != "debian" {
		t.Errorf("Expected distribution \"%s\"; got \"%s\"", "debian", distribution)
	}
	if release != "8" {
		t.Errorf("Expected release \"%s\"; got \"%s\"", "8", release)
	}

	// 3.2: unknown
	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Err: stubutils.ErrOhNo},
		&stubutils.CmdStub{Cmd: "lsb_release", Args: []string{"-ir"}, Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/issue", Output: []byte("unknown 8 \n \\l")},
		&stubutils.ReadFileStub{Path: "/etc/centos-release", Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/redhat-release", Output: []byte("unknown 8 \n \\l")},
		&stubutils.ReadFileStub{Path: "/etc/SuSE-release", Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/system-release", Output: []byte("unknown 8 \n \\l")},
		&stubutils.ReadFileStub{Path: "/etc/system-release-cpe", Err: stubutils.ErrOhNo},
	)

	_, _, err = GetDistributionRelease()
	if err != ErrUnknownDistribution {
		t.Errorf("Expected error \"%s\"; got \"%s\"", ErrUnknownDistribution.Error(), err.Error())
	}
}

func TestGetFqdn(t *testing.T) {
	s := stubutils.NewStubs(t)
	defer s.Close()

	// 1: all good
	s.Add(&stubutils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Output: []byte("HOST")})
	if out, _ := GetFqdn(); out != "HOST" {
		t.Errorf("Expected \"host\"; got %s", out)
	}

	// 2: new line
	s.Add(&stubutils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Output: []byte("HOST\n")})
	if out, _ := GetFqdn(); out != "HOST" {
		t.Errorf("Expected \"host\"; got %s", out)
	}

	// 3: file
	s.Add(&stubutils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/hostname", Output: []byte("HOST")})
	if out, _ := GetFqdn(); out != "HOST" {
		t.Errorf("Expected \"host\"; got %s", out)
	}

	// 4: unknown
	s.Add(&stubutils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Err: stubutils.ErrOhNo},
		&stubutils.ReadFileStub{Path: "/etc/hostname", Err: stubutils.ErrOhNo})
	if _, err := GetFqdn(); err != ErrUnknownFqdn {
		t.Errorf("Expected error \"%s\"; got %s", ErrUnknownFqdn.Error(), err.Error())
	}

}

func TestGetArchitecture(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Output: []byte("ARCH")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Output: []byte("ARCH\n")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Err: stubutils.ErrOhNo})
	defer s.Close()

	// 1: all good
	if out, _ := GetArchitecture(); out != "ARCH" {
		t.Errorf("Expected \"arch\"; got %s", out)
	}

	// 2: new line
	if out, _ := GetArchitecture(); out != "ARCH" {
		t.Errorf("Expected \"arch\"; got %s", out)
	}

	// 3:  unknown
	if _, err := GetArchitecture(); err != ErrUnknownArchitecture {
		t.Errorf("Expected error \"%s\"; got %s", ErrUnknownArchitecture.Error(), err.Error())
	}

}

func TestGetKernel(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Output: []byte("KERN")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Output: []byte("KERN\n")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Err: stubutils.ErrOhNo})
	defer s.Close()

	// 1: all good
	if out, _ := GetKernel(); out != "KERN" {
		t.Errorf("Expected \"kern\"; got %s", out)
	}

	// 2: new line
	if out, _ := GetKernel(); out != "KERN" {
		t.Errorf("Expected \"kern\"; got %s", out)
	}

	// 3:  unknown
	if _, err := GetKernel(); err != ErrUnknownKernel {
		t.Errorf("Expected error \"%s\"; got %s", ErrUnknownKernel.Error(), err.Error())
	}

}
