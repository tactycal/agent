// +build rhel centos debian

package main

import "testing"

func TestGetHostRelease(t *testing.T) {
	fixtures := []struct {
		title    string
		stubs    []ioStub
		expected string
	}{
		{
			title: "rhel os-release (rhel 7.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"7.3\"")},
			},
			expected: "7",
		},
		{
			title: "rhel lsb_release (rhel 7.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release    7.3")},
			},
			expected: "7",
		},
		{
			title: "rhel redhat-release (rhel 7.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", err: ohNoErr},
				&readFileStub{path: "/etc/redhat-release", output: []byte("Red Hat Enterprise Linux Server release 7.3 (Maipo)")},
			},
			expected: "7",
		},
		{
			title: "centos os-release (centos 7.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"7\"")},
			},
			expected: "7",
		},
		{
			title: "centos lsb_release (centos 7.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release    7.3.1611")},
			},
			expected: "7",
		},
		{
			title: "centos redhat-release (centos 7.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", output: []byte("CentOS Linux release 7.3.1611 (Core)")},
			},
			expected: "7",
		},
		{
			title: "debian os-release (debian 8.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"8\"")},
			},
			expected: "8",
		},
		{
			title: "debian lsb_release (debian 8.3)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release    8.3")},
			},
			expected: "8",
		},
	}

	for _, fix := range fixtures {
		t.Run(fix.title, func(t *testing.T) {
			s := newStubs(t, fix.stubs...)
			defer s.Close()

			if out := getHostRelease(); out != fix.expected {
				t.Errorf("Expected \"%s\", got \"%s\"", fix.expected, out)
			}
		})
	}
}
