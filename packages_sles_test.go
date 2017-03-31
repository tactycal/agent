// +build sles

package main

import (
	"reflect"
	"testing"
)

func TestBuildPckage(t *testing.T) {
	testCase := map[string]string{
		"Name":    "libtasn1",
		"Arch":    "x86_64",
		"Version": "3.7",
		"Release": "12.2",
		"Source":  "libtasn1-3.7-12.2.src.rpm",
		"Vendor":  "SUSE LLC <https://www.suse.com/>",
		"Epoch":   "(none)",
	}

	expectedResult := &Package{
		Name:         "libtasn1",
		Version:      "3.7-12.2",
		Architecture: "x86_64",
		Official:     true,
		Source:       "libtasn1",
	}

	result := buildPackage(testCase)
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}

func TestReadPackages(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "rpm", args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArch: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			stubFile: "testdata/sles_rpm"})
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

	result, _ := readPackages()

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}
