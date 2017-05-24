package osDiscovery

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetValues(t *testing.T) {
	content := `NAME="Ubuntu"
VERSION="14.04.5 LTS, Trusty Tahr"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 14.04.5 LTS"
VERSION_ID="14.04"
HOME_URL=
Distributor ID: CentOS
Description:    CentOS Linux release 7.3.1611 (Core)
Release:    7.3.1611
Codename:   Core
VERSION = 12
PATCHLEVEL = 2
`
	testCases := []struct {
		title          string
		regexpField1   string
		regexpField2   string
		expectedField1 string
		expectedField2 string
	}{
		{"1", "VERSION_ID=", "ID=", "14.04", "ubuntu"},
		{"2", "Distributor ID:", "Release:", "CentOS", "7.3.1611"},
		{"3", "VERSION =", "PATCHLEVEL =", "12", "2"},
		{"4", "VERSION =", "Test=", "12", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			field1, field2 := getValues(testCase.regexpField1, testCase.regexpField2, content)
			if field1 != testCase.expectedField1 {
				t.Errorf("Expected field1 \"%s\"; got \"%s\"", testCase.expectedField1, field1)
			}
			if field2 != testCase.expectedField2 {
				t.Errorf("Expected field2 \"%s\"; got \"%s\"", testCase.expectedField2, field2)
			}
		})
	}
}

func TestDistributionReleaseSpecific(t *testing.T) {
	testCases := []struct {
		Title                string `json:"title"`
		Content              string `json:"content"`
		ExpectedDistribution string `json:"expectedDistribution"`
		ExpectedRelease      string `json:"expectedRelease"`
	}{}

	testFile, _ := os.Open("testdata/distroSpecific.json")
	err := json.NewDecoder(testFile).Decode(&testCases)
	if err != nil {
		t.Errorf("Error not expected; got %s", err.Error())
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			distribution, release, _ := parseSpecificFileContent(testCase.Content)
			if distribution != testCase.ExpectedDistribution {
				t.Errorf("Expected distribution \"%s\"; got \"%s\"", testCase.ExpectedDistribution, distribution)
			}
			if release != testCase.ExpectedRelease {
				t.Errorf("Expected release \"%s\"; got \"%s\"", testCase.ExpectedRelease, release)
			}
		})
	}
}

func TestGetReleaseFromSpecificFileContentDefault(t *testing.T) {
	testCases := []struct {
		title    string
		content  string
		regexp   string
		expected string
	}{
		{"1", "xxx xxxx 7.3.456 xx 4.56 xx", "\\d+", "7"},
		{"2", "xxx xxxx 7.3.456 xx 4.56 xx", "\\d+\\.\\d+", "7.3"},
		{"3", "xxx xxxx 7.34.456 xx 4.56 xx", "\\d+\\.\\d+", "7.34"},
		{"4", "xxx xxxx 7. xxx", "\\d+\\.\\d+", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			release := getReleaseFromSpecificFileContentDefault(testCase.regexp, testCase.content)
			if release != testCase.expected {
				t.Errorf("Expected release \"%s\"; got \"%s\"", testCase.expected, release)
			}
		})
	}
}

func TestParseLsbReleaseContent(t *testing.T) {
	testCases := []struct {
		Title                string `json:"title"`
		Content              string `json:"content"`
		ExpectedDistribution string `json:"expectedDistribution"`
		ExpectedRelease      string `json:"expectedRelease"`
	}{}

	testFile, _ := os.Open("testdata/lsbFallback.json")
	err := json.NewDecoder(testFile).Decode(&testCases)
	if err != nil {
		t.Errorf("Error not expected; got %s", err.Error())
	}

	for _, testCase := range testCases {
		t.Run(testCase.Title, func(t *testing.T) {
			distribution, release, _ := parseLsbReleaseContent(testCase.Content)
			if distribution != testCase.ExpectedDistribution {
				t.Errorf("Expected distribution \"%s\"; got \"%s\"", testCase.ExpectedDistribution, distribution)
			}
			if release != testCase.ExpectedRelease {
				t.Errorf("Expected release \"%s\"; got \"%s\"", testCase.ExpectedRelease, release)
			}
		})
	}
}
