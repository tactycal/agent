// +build ubuntu opensuse amzn

package main

import "testing"

func TestGetHostRelease(t *testing.T) {
	fixtures := []struct {
		title    string
		stubs    []ioStub
		expected string
	}{
		{
			title: "ubuntu os-release (ubuntu 14.04)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"14.04\"")},
			},
			expected: "14.04",
		},
		{
			title: "ubuntu lsb_release (ubuntu 14.04)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release    14.04")},
			},
			expected: "14.04",
		},
		{
			title: "opensuse os-release (opensuse 42.2)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"42.2\"")},
			},
			expected: "42.2",
		},
		{
			title: "opensuse lsb_release (opensuse 42.2)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release:    42.2")},
			},
			expected: "42.2",
		},
		{
			title: "opensuse SuSE-release (opensuse 42.2)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", err: ohNoErr},
				&readFileStub{path: "/etc/redhat-release", err: ohNoErr},
				&readFileStub{path: "/etc/SuSE-release", output: []byte("openSUSE 42.2 (x86_64)\nVERSION = 42.2\nCODENAME = Malachite\n")},
			},
			expected: "42.2",
		},
		{
			title: "amzn os-release (amzn 2017.03)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"2017.03\"")},
			},
			expected: "2017.03",
		},
		{
			title: "amzn lsb_release (amzn 2017.03)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release:        2017.03")},
			},
			expected: "2017.03",
		},
		{
			title: "amzn system-release (amzn 2017.03)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", err: ohNoErr},
				&readFileStub{path: "/etc/redhat-release", err: ohNoErr},
				&readFileStub{path: "/etc/SuSE-release", err: ohNoErr},
				&readFileStub{path: "/etc/system-release", output: []byte("Amazon Linux AMI release 2017.03")},
			},
			expected: "2017.03",
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
