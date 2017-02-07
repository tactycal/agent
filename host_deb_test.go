// +build ubuntu debian

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
			title: "ubuntu os-release (debian 8)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"8\"")},
			},
			expected: "8",
		},
		{
			title: "lsb_release (ubuntu 14.04.5)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", output: []byte("Ubuntu 14.04.5 LTS")},
			},
			expected: "14.04",
		},
		{
			title: "lsb_release (ubuntu 12.04)",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", output: []byte("Ubuntu 12.04 LTS")},
			},
			expected: "12.04",
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
