// +build sles

package main

import "testing"

func TestGetHostRelease(t *testing.T) {
	fixtures := []struct {
		title    string
		stubs    []ioStub
		expected string
	}{
		{
			title: "sles os-release (sles 12)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"12\"")},
			},
			expected: "12",
		},
		{
			title: "sles os-release (sles 12.2)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"12.2\"")},
			},
			expected: "12.2",
		},
		{
			title: "sles lsb_release (sles 12)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release:      12")},
			},
			expected: "12",
		},
		{
			title: "sles lsb_release (sles 12.2)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, output: []byte("Release:      12.2")},
			},
			expected: "12.2",
		},
		{
			title: "sles SuSE-release (sles 12)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", err: ohNoErr},
				&readFileStub{path: "/etc/redhat-release", err: ohNoErr},
				&readFileStub{path: "/etc/SuSE-release", output: []byte("SUSE Linux Enterprise Server 12 (x86_64)\nVERSION = 12\nPATCHLEVEL = 0")},
			},
			expected: "12",
		},
		{
			title: "sles SuSE-release (sles 12.2)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-r"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", err: ohNoErr},
				&readFileStub{path: "/etc/redhat-release", err: ohNoErr},
				&readFileStub{path: "/etc/SuSE-release", output: []byte("SUSE Linux Enterprise Server 12 (x86_64)\nVERSION = 12\nPATCHLEVEL = 2")},
			},
			expected: "12.2",
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
