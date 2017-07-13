package main

import (
	"reflect"
	"testing"

	"github.com/tactycal/agent/stubUtils"
)

func TestLocal(t *testing.T) {
	s := stubUtils.NewStubs(t,
		&stubUtils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=amzn\nVERSION_ID=\"2016.09\"")},
		&stubUtils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Output: []byte("ARCH")},
		&stubUtils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Output: []byte("KERN")},
		&stubUtils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Output: []byte("FQDN")},
		&stubUtils.CmdStub{Cmd: "rpm", Args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			Output: []byte("Name: libverto\nArchitecture: x86_64\nVersion: 0.2.5\nRelease: 4.9\nVendor: unknown\nSource: libverto-0.2.5-4.9.src.rpm\nEpoch: (none)")},
	)
	defer s.Close()

	expected := `{"fqdn":"FQDN","distribution":"amzn","release":"2016.09","architecture":"ARCH","kernel":"KERN","labels":null,"packages":[{"name":"libverto","version":"0.2.5-4.9","source":"libverto","architecture":"x86_64","official":false}]}`

	// 1. ok
	result, _ := local()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Result\n%s\ndoesn't match expected\n%s\n", result, expected)
	}

	// 2. error expected from get host info
	s.Add(&stubUtils.ReadFileStub{Path: "/etc/os-release", Err: stubUtils.OhNoErr})

	_, err := local()
	if err == nil {
		t.Error("An error was expected")
	}

	// 3. error is expected from packageLookup
	s.Add(&stubUtils.ReadFileStub{Path: "/etc/os-release", Output: []byte("ID=amzn\nVERSION_ID=\"2016.09\"")},
		&stubUtils.CmdStub{Cmd: "uname", Args: []string{"-m"}, Output: []byte("ARCH")},
		&stubUtils.CmdStub{Cmd: "uname", Args: []string{"-r"}, Output: []byte("KERN")},
		&stubUtils.CmdStub{Cmd: "hostname", Args: []string{"-f"}, Output: []byte("FQDN")},
		&stubUtils.CmdStub{Cmd: "rpm", Args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			Err: stubUtils.OhNoErr},
	)

	_, err = local()
	if err == nil {
		t.Error("An error was expected")
	}
}
