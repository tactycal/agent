// +build  ubuntu

package main

import (
	"reflect"
	"testing"
)

func TestGetPackages(t *testing.T) {
	s := newStubs(t,
		&readFileStub{path: dpkgStatusPath, stubFile: "testdata/ubuntu_status"},
		&readFileStub{path: sourcesListPath, stubFile: "testdata/ubuntu_source"},
		&cmdStub{cmd: "apt-cache", stubFile: "testdata/ubuntu_apt_cache"})
	defer s.Close()

	expectedResult := map[string]*Package{
		"gtk2-engines-murrine": &Package{
			Name:         "gtk2-engines-murrine",
			Version:      "0.98.2-0ubuntu1",
			Architecture: "i386",
			Maintainer:   "Ubuntu Core Developers <ubuntu-devel-discuss@lists.ubuntu.com>",
			Official:     true,
		},
		"skype": &Package{
			Name:         "skype",
			Version:      "4.2.0.11-1",
			Architecture: "i386",
			Maintainer:   "Skype Technologies <info@skype.net>",
			Official:     false,
		},
		"apt": &Package{
			Name:         "apt",
			Version:      "0.8.16~exp12ubuntu10.27",
			Maintainer:   "Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>",
			Architecture: "amd64",
			Official:     true,
		},
		"oracle-java8-installer": &Package{
			Name:         "oracle-java8-installer",
			Version:      "8u111+8u111arm-1~webupd8~0",
			Maintainer:   "Alin Andrei <webupd8@gmail.com>",
			Architecture: "all",
			Official:     false,
		},
		"apt-xapian-index": &Package{
			Name:         "apt-xapian-index",
			Version:      "0.44ubuntu5.1",
			Maintainer:   "Ubuntu Developers <ubuntu-devel-discuss@lists.ubuntu.com>",
			Architecture: "all",
			Official:     true,
		},
	}

	result, err := GetPackages()

	// check error
	if err != nil {
		t.Errorf("Expected error to bi nil; got %v", err)
	}

	// check if numbers of packages match
	if len(result) != len(expectedResult) {
		t.Errorf("Number of packages %d doesn't match expected %d", len(result), len(expectedResult))
	}

	// check if packages match
	for _, v := range result {
		// check if expected result contain package
		expectedV, ok := expectedResult[v.Name]
		if !ok {
			t.Errorf("Package %s is not expected", v.Name)
			continue
		}

		if !reflect.DeepEqual(expectedV, v) {
			t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", v, expectedV)
		}
	}
}

func TestExtractPackageNameFromSource(t *testing.T) {
	testCases := []struct {
		title          string
		source         string
		expectedResult string
	}{
		{"1", "gtk+3.0", "gtk+3.0"},
		{"2", "libsoup2.4", "libsoup2.4"},
		{"3", "sane-backends", "sane-backends"},
		{"4", "libxbc", "libxbc"},
		{"5", "lvm2 (2.02.66-4ubuntu7.4)", "lvm2"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			result := extractPackageNameFromSource(testCase.source)
			if result != testCase.expectedResult {
				t.Errorf("Expected %s got %s", testCase.expectedResult, result)
			}
		})
	}

}

func TestGetRepositoriesFromSourcesList(t *testing.T) {
	s := newStubs(t,
		&readFileStub{path: sourcesListPath, stubFile: "testdata/ubuntu_source"})
	defer s.Close()

	expectedResult := []string{
		"http://archive.ubuntu.com/ubuntu",
	}

	result, err := getRepositoriesFromSourcesList()
	if err != nil {
		t.Errorf("An error was not expected; err = %s", err.Error())
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}

func TestOfficialMapToList(t *testing.T) {
	testCase := map[string]*Package{
		"Package": &Package{
			Name: "Package",
		},
	}

	expectedResult := []string{"Package"}

	result := officialMapToList(testCase)

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", result, expectedResult)
	}
}

func TestGetAptCachePolicy(t *testing.T) {
	s := newStubs(t,
		&cmdStub{cmd: "apt-cache", stubFile: "testdata/ubuntu_apt_cache"}, // 0.1
		&cmdStub{cmd: "apt-cache", err: ohNoErr})                          // 0.2
	defer s.Close()

	testCase := []string{
		"gtk2-engines-murrine",
		"skype",
		"apt",
		"oracle-java8-installer",
		"apt-xapian-index",
	}

	// 0.1
	expectedResult := map[string][]string{
		"gtk2-engines-murrine": []string{},
		"skype:i386":           []string{},
		"apt": []string{
			"http://archive.ubuntu.com/ubuntu",
		},
		"oracle-java8-installer": []string{
			"http://ppa.launchpad.net/webupd8team/java/ubuntu",
		},
		"apt-xapian-index": []string{
			"http://archive.ubuntu.com/ubuntu",
			"http://archive.ubuntu.com/ubuntu",
		},
	}

	result, err := getAptCachePolicy(testCase)

	// check error
	if err != nil {
		t.Fatalf("Expected error to be nil; got %v", err)
	}

	// check a map length
	if len(result) != len(expectedResult) {
		t.Errorf("Number of keys in result %d doesn't match expected %d", len(result), len(expectedResult))
	}

	// check if values match
	for k, v := range expectedResult {
		if !reflect.DeepEqual(result[k], v) {
			t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n; for key %s", result[k], v, k)
		}
	}

	// 0.2 expected error from apt-cache
	result, err = getAptCachePolicy(testCase)

	// check error
	if err == nil {
		t.Error("An error was expected")
	}

	// check result
	if result != nil {
		t.Errorf("Expected result to be nil; got %+v", result)
	}
}

func TestIsPackageSourceFromOfficialRepositories(t *testing.T) {
	testCases := []struct {
		title          string
		sources        []string
		officialRepos  []string
		expectedResult bool
	}{
		{
			"only official",
			[]string{"http://archive.ubuntu.com/ubuntu"},
			[]string{"http://archive.ubuntu.com/ubuntu"},
			true,
		},
		{
			"official and unofficial",
			[]string{
				"http://ppa.launchpad.net/webupd8team/java/ubuntu",
				"http://archive.ubuntu.com/ubuntu",
			},
			[]string{"http://archive.ubuntu.com/ubuntu"},
			true,
		},
		{
			"unofficial",
			[]string{"http://ppa.launchpad.net/webupd8team/java/ubuntu"},
			[]string{"http://archive.ubuntu.com/ubuntu"},
			false,
		},
		{
			"empty sources",
			[]string{},
			[]string{"http://archive.ubuntu.com/ubuntu"},
			false,
		},
		{
			"empty officialRepos",
			[]string{"http://archive.ubuntu.com/ubuntu"},
			[]string{},
			false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			result := isPackageSourceFromOfficialRepositories(testCase.sources, testCase.officialRepos)
			if result != testCase.expectedResult {
				t.Errorf("Expected %t got %t", testCase.expectedResult, result)
			}
		})
	}
}

func TestSetOfficial(t *testing.T) {
	s := newStubs(t,
		&readFileStub{path: sourcesListPath, stubFile: "testdata/ubuntu_source"},
		&cmdStub{cmd: "apt-cache", stubFile: "testdata/ubuntu_apt_cache"})
	defer s.Close()

	testCase := map[string]*Package{
		"gtk2-engines-murrine": &Package{
			Name:         "gtk2-engines-murrine",
			Version:      "0.98.2-0ubuntu1",
			Architecture: "i386",
			Maintainer:   "Ubuntu Core Developers <ubuntu-devel-discuss@lists.ubuntu.com>",
		},
		"skype": &Package{
			Name:         "skype",
			Version:      "4.2.0.11-1",
			Architecture: "i386",
			Maintainer:   "Skype Technologies <info@skype.net>",
		},
	}

	expectedResult := map[string]*Package{
		"gtk2-engines-murrine": &Package{
			Name:         "gtk2-engines-murrine",
			Version:      "0.98.2-0ubuntu1",
			Architecture: "i386",
			Maintainer:   "Ubuntu Core Developers <ubuntu-devel-discuss@lists.ubuntu.com>",
			Official:     true,
		},
		"skype": &Package{
			Name:         "skype",
			Version:      "4.2.0.11-1",
			Architecture: "i386",
			Maintainer:   "Skype Technologies <info@skype.net>",
			Official:     false,
		},
	}

	setOfficial(testCase)

	if !reflect.DeepEqual(testCase, expectedResult) {
		t.Errorf("Result\n%+v\ndoesn't match expected\n%+v\n", testCase, expectedResult)
	}
}
