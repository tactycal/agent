// +build rhel

package main

import (
	"reflect"
	"testing"
)

func TestBuildVersion(t *testing.T) {
	testCases := []struct {
		title          string
		matches        map[string]string
		expectedResult string
	}{
		{
			"with epoch",
			map[string]string{
				"Epoch":   "1",
				"Version": "0.9.9.1",
				"Release": "13.git20140326.4dba720.el7",
			},
			"1:0.9.9.1-13.git20140326.4dba720.el7",
		},
		{
			"without epoch",
			map[string]string{
				"Version": "0",
				"Release": "2.el7",
			},
			"0-2.el7",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			result := buildVersion(testCase.matches)
			if result != testCase.expectedResult {
				t.Errorf("Expected %s got %s", testCase.expectedResult, result)
			}
		})
	}
}

func TestBuildPckage(t *testing.T) {
	testCase := map[string]string{
		"Name":    "NetworkManager-config-server",
		"Arch":    "x86_64",
		"Epoch":   "1",
		"Version": "0.9.9.1",
		"Release": "13.git20140326.4dba720.el7",
		"Vendor":  "local",
	}

	expectedResult := &Package{
		Name:         "NetworkManager-config-server",
		Version:      "1:0.9.9.1-13.git20140326.4dba720.el7",
		Architecture: "x86_64",
		Official:     false,
		Source:       "unknown",
	}

	result := buildPackage(testCase)
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}

func TestReadPackages(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "rpm", args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArch: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`}, stubFile: "testdata/rhel_rpm"}, // 0.1
		&cmdStub{err: ohNoErr}) // 0.2
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

	result, err := readPackages()

	// check error
	if err != nil {
		t.Errorf("Expected error to be nil; got %v", err)
	}

	// check result
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}

	// 0.2 expected error from rpm
	result, err = readPackages()

	// check error
	if err == nil {
		t.Error("An error was expected")
	}

	// check result
	if len(result) > 0 {
		t.Errorf("Result was expected to be empty; got %+v", result)
	}
}
