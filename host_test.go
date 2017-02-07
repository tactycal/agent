package main

import (
	"reflect"
	"testing"
)

func TestGetHostInfo(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "hostname", args: []string{"-f"}, output: []byte("HOST")},
		&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"8\"")},
		&cmdStub{cmd: "uname", args: []string{"-m"}, output: []byte("ARCH")},
		&cmdStub{cmd: "uname", args: []string{"-r"}, output: []byte("KERN")})
	defer s.Close()

	expected := Host{
		Fqdn:         "HOST",
		Distribution: DISTRIBUTION,
		Release:      "8",
		Architecture: "ARCH",
		Kernel:       "KERN",
	}

	if out := GetHostInfo(); !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected\n%+v to equal\n%+v", expected, out)
	}
}

func TestGetHostFqdn(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "hostname", args: []string{"-f"}, output: []byte("HOST")},
		&cmdStub{cmd: "hostname", args: []string{"-f"}, err: ohNoErr})
	defer s.Close()

	// 01: all good
	if out := getHostFqdn(); out != "HOST" {
		t.Errorf("Expected \"host\"; got %s", out)
	}

	// 02: set default
	if out := getHostFqdn(); out != "unknown" {
		t.Errorf("Expected \"unknown\"; got %s", out)
	}
}

func TestGetHostArchitecture(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "uname", args: []string{"-m"}, output: []byte("ARCH")},
		&cmdStub{cmd: "uname", args: []string{"-m"}, err: ohNoErr})
	defer s.Close()

	// 01: all good
	if out := getHostArchitecture(); out != "ARCH" {
		t.Errorf("Expected \"arch\"; got %s", out)
	}

	// 02: set default
	if out := getHostArchitecture(); out != "unknown" {
		t.Errorf("Expected \"default\"; got %s", out)
	}
}

func TestGetHostKernel(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "uname", args: []string{"-r"}, output: []byte("KERN")},
		&cmdStub{cmd: "uname", args: []string{"-r"}, err: ohNoErr})
	defer s.Close()

	// 01: all good
	if out := getHostKernel(); out != "KERN" {
		t.Errorf("Expected \"arch\"; got %s", out)
	}

	// 02: set default
	if out := getHostKernel(); out != "unknown" {
		t.Errorf("Expected \"default\"; got %s", out)
	}
}

func TestReadHostRelease(t *testing.T) {
	fixtures := []struct {
		title  string
		stubs  []ioStub
		output string
	}{
		{
			title: "os-release",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte("VERSION_ID=\"OS-RELEASE\"")},
			},
			output: "OS-RELEASE",
		},
		{
			title: "lsb_release",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-sd"}, output: []byte("LSB_RELEASE")},
			},
			output: "LSB_RELEASE",
		},
		{
			title: "centos_release",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", output: []byte{}},
				&cmdStub{cmd: "lsb_release", args: []string{"-sd"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", output: []byte("CENTOS_RELEASE")},
			},
			output: "CENTOS_RELEASE",
		},
		{
			title: "redhat_release",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-sd"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", err: ohNoErr},
				&readFileStub{path: "/etc/redhat-release", output: []byte("REDHAT_RELEASE")},
			},
			output: "REDHAT_RELEASE",
		},
		{
			title: "default",
			stubs: []ioStub{
				&readFileStub{path: "/etc/os-release", err: ohNoErr},
				&cmdStub{cmd: "lsb_release", args: []string{"-sd"}, err: ohNoErr},
				&readFileStub{path: "/etc/centos-release", err: ohNoErr},
				&readFileStub{path: "/etc/redhat-release", err: ohNoErr},
			},
			output: "unknown",
		},
	}

	for _, fix := range fixtures {
		t.Run(fix.title, func(t *testing.T) {
			s := newStubs(t, fix.stubs...)
			defer s.Close()

			if out := readHostRelease(); out != fix.output {
				t.Errorf("Expected \"%s\"; got \"%s\"", fix.output, out)
			}
		})
	}
}
