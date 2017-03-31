// +build amzn

package main

import (
	"reflect"
	"testing"
)

func TestReadPackages(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "rpm", args: []string{`-qa`, `--queryformat`,
			`Name: %{NAME}\nArch: %{ARCH}\nVersion: %{VERSION}\nRelease: %{RELEASE}\nVendor: %{VENDOR}\nSource: %{SOURCERPM}\nEpoch: %{EPOCH}\n\n`},
			stubFile: "testdata/amzn_rpm"})
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

	result, _ := readPackages()

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}
