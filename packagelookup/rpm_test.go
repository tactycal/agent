package packagelookup

import (
	"reflect"
	"testing"

	"github.com/tactycal/agent/stubutils"
)

func TestGetSourceName(t *testing.T) {
	testCases := []struct {
		source       map[string]string
		expectedName string
	}{
		{map[string]string{"Source": "perl-Sys-Syslog-0.33-3.el7.src.rpm"}, "perl-Sys-Syslog"},
		{map[string]string{"Source": "ibtasn1-3.7-12.2.src.rpm"}, "ibtasn1"},
		{map[string]string{"Source": "insserv-compat-0.1-15.55.src.rpm"}, "insserv-compat"},
		{map[string]string{"Source": "wrong-version.src.rpm"}, "unknown"},
	}

	for _, testCase := range testCases {
		testName := testCase.expectedName
		t.Run(testName, func(t *testing.T) {
			name := getSourceName(testCase.source)
			if name != testCase.expectedName {
				t.Errorf("Expected \"%s\", got \"%s\"", testCase.expectedName, name)
			}
		})
	}
}

func TestGet_amzn(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.CmdStub{Cmd: "rpm", Args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			StubFile: "testdata/amzn_rpm"})
	defer s.Close()

	expectedResult := []*Package{
		&Package{
			Name:         "make",
			Version:      "1:3.82-21.10.amzn1",
			Architecture: "x86_64",
			Official:     true,
			Source:       "make",
		},
		&Package{
			Name:         "libverto",
			Version:      "0.2.5-4.9",
			Architecture: "x86_64",
			Official:     false,
			Source:       "libverto",
		},
	}

	result, _ := Get("amzn")

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}

}

func TestGet_opensuse(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.CmdStub{Cmd: "rpm", Args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			StubFile: "testdata/opensuse_rpm"})
	defer s.Close()

	expectedResult := []*Package{
		&Package{
			Name:         "libtasn1",
			Version:      "3.7-12.2",
			Architecture: "x86_64",
			Official:     true,
			Source:       "libtasn1",
		},
		&Package{
			Name:         "logrotate",
			Version:      "3.8.7-8.4",
			Architecture: "x86_64",
			Official:     false,
			Source:       "logrotate",
		},
	}

	result, _ := Get("opensuse")
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}

func TestGet_sles(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.CmdStub{Cmd: "rpm", Args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			StubFile: "testdata/sles_rpm"})
	defer s.Close()

	expectedResult := []*Package{
		&Package{
			Name:         "libtasn1",
			Version:      "3.7-12.2",
			Architecture: "x86_64",
			Official:     true,
			Source:       "libtasn1",
		},
		&Package{
			Name:         "logrotate",
			Version:      "3.8.7-8.4",
			Architecture: "x86_64",
			Official:     false,
			Source:       "logrotate",
		},
	}

	result, _ := Get("sles")

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}

func TestGet_rhel(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.CmdStub{Cmd: "rpm", Args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`}, StubFile: "testdata/rhel_rpm"}, // 0.1
		&stubutils.CmdStub{Err: stubutils.ErrOhNo}) // 0.2
	defer s.Close()

	// 0.1
	expectedResult := []*Package{
		&Package{
			Name:         "NetworkManager-config-server",
			Version:      "1:0.9.9.1-13.git20140326.4dba720.el7",
			Architecture: "x86_64",
			Official:     false,
			Source:       "NetworkManager",
		},
		&Package{
			Name:         "Red_Hat_Enterprise_Linux-Release_Notes-7-en-US",
			Architecture: "noarch",
			Version:      "0-2.el7",
			Official:     true,
			Source:       "Red_Hat_Enterprise_Linux-Release_Notes-7-en-US",
		},
	}

	result, err := Get("rhel")

	// check error
	if err != nil {
		t.Errorf("Expected error to be nil; got %v", err)
	}

	// check result
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}

	// 0.2 expected error from rpm
	result, err = Get("rhel")

	// check error
	if err == nil {
		t.Error("An error was expected")
	}

	// check result
	if len(result) > 0 {
		t.Errorf("Result was expected to be empty; got %+v", result)
	}
}

func TestGet_centos(t *testing.T) {
	s := stubutils.NewStubs(t,
		&stubutils.CmdStub{Cmd: "rpm", Args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArchitecture: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			StubFile: "testdata/centos_rpm"})
	defer s.Close()

	expectedResult := []*Package{
		&Package{
			Name:         "audit-libs",
			Version:      "2.3.7-5.el6",
			Architecture: "x86_64",
			Official:     true,
			Source:       "audit",
		},
	}

	result, _ := Get("centos")

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}
