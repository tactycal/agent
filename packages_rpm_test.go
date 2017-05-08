// +build sles opensuse rhel centos

package main

import "testing"

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
