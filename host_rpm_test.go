// +build rhel centos

package main

import "testing"

func TestGetHostRelease(t *testing.T) {
	s := newStubs(t,
		&readFileStub{path: "/etc/os-release", err: ohNoErr},
		&cmdStub{cmd: "lsb_release", args: []string{"-sd"}, err: ohNoErr},
		&readFileStub{path: "/etc/centos-release", output: []byte(" CentOS Linux release 7.2.1511 (Core) ")})
	defer s.Close()

	if out := getHostRelease(); out != "7" {
		t.Errorf("Expected \"RELEASE\"; got \"%s\"", out)
	}
}
