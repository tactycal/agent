// +build centos

package main

import (
	"reflect"
	"testing"
)

func TestBuildPckage(t *testing.T) {
	testCase := map[string]string{
		"Name":    "audit-libs",
		"Arch":    "x86_64",
		"Version": "2.3.7",
		"Release": "5.el6",
		"Vendor":  "CentOS",
	}

	expectedResult := &Package{
		Name:         "audit-libs",
		Version:      "2.3.7-5.el6",
		Architecture: "x86_64",
		Official:     true,
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
			`Name: %{NAME}\nArch: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			stubFile: "testdata/centos_rpm"})
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

	result, _ := readPackages()

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}
