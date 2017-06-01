package main

import (
	"reflect"
	"testing"

	"github.com/tactycal/agent/stubUtils"
)

func TestGetHostInfo(t *testing.T) {
	s := stubUtils.NewStubs(t,
		&stubUtils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=ubuntu\nID_LIKE=debian\nVERSION_ID=\"14.04\"")},
		&stubUtils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Output: []byte("ARCH")},
		&stubUtils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Output: []byte("KERN")},
		&stubUtils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Output: []byte("FQDN")},
	)
	defer s.Close()
	expected := &Host{
		Fqdn:         "FQDN",
		Distribution: "ubuntu",
		Release:      "14.04",
		Architecture: "ARCH",
		Kernel:       "KERN",
	}

	host, _ := GetHostInfo()
	if !reflect.DeepEqual(host, expected) {
		t.Errorf("Host\n%+v\ndoesn't match expected\n%+v\n", host, expected)
	}

	s.Add(&stubUtils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=ubuntu\nID_LIKE=debian\nVERSION_ID=\"14.04\"")},
		&stubUtils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Err: stubUtils.OhNoErr})

	_, err := GetHostInfo()
	if err == nil {
		t.Error("An error was expected")
	}
}
