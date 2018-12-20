package main

import (
	"reflect"
	"testing"

	"github.com/tactycal/agent/stubutils"
)

func TestGetHostInfo(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=ubuntu\nID_LIKE=debian\nVERSION_ID=\"14.04\"")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Output: []byte("ARCH")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Output: []byte("KERN")},
		&stubutils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Output: []byte("FQDN")},
	)
	defer s.Close()
	expected := &Host{
		Fqdn:         "FQDN",
		Distribution: "ubuntu",
		Release:      "14.04",
		Architecture: "ARCH",
		Kernel:       "KERN",
	}

	host, _ := getHostInfo()
	if !reflect.DeepEqual(host, expected) {
		t.Errorf("Host\n%+v\ndoesn't match expected\n%+v\n", host, expected)
	}

	s.Add(&stubutils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=ubuntu\nID_LIKE=debian\nVERSION_ID=\"14.04\"")},
		&stubutils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Err: stubutils.ErrOhNo})

	_, err := getHostInfo()
	if err == nil {
		t.Error("An error was expected")
	}
}
