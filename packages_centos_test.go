// +build centos

package main

import (
	"reflect"
	"testing"
)

func TestIsOfficial(t *testing.T) {
	testCases := []struct {
		title          string
		fromRepo       string
		expectedResult bool
	}{
		{"from CentOS", "CentOS", true},
		{"from Updates", "Updates", true},
		{"from base", "base", true},
		{"from unkown", "unkown", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			result := isOfficial(testCase.fromRepo)
			if result != testCase.expectedResult {
				t.Errorf("Expected %t got %t", testCase.expectedResult, result)
			}
		})
	}
}

func TestBuildPckage(t *testing.T) {
	testCase := map[string]string{
		"Name":      "audit-libs",
		"Arch":      "x86_64",
		"Version":   "2.3.7",
		"Release":   "5.el6",
		"From repo": "CentOS",
	}

	expectedResult := &Package{
		Name:         "audit-libs",
		Version:      "2.3.7-5.el6",
		Architecture: "x86_64",
		Official:     true,
	}

	result := buildPackage(testCase)
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}

func TestReadPackages(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "yum", args: []string{"info", "installed"}, stubFile: "testdata/centos_yum"})
	defer s.Close()

	expectedResult := []*Package{
		&Package{
			Name:         "audit-libs",
			Version:      "2.3.7-5.el6",
			Architecture: "x86_64",
			Official:     true,
		},
	}

	result, _ := readPackages()

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}
