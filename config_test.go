package main

import (
	"bytes"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tactycal/agent/stubutils"
)

func TestNewConfig_Valid(t *testing.T) {
	fixtures := []struct {
		title    string
		data     []byte
		expected *config
	}{
		{
			title: "minimal config",
			data:  []byte(`token=TOKEN`),
			expected: &config{
				Token:         "TOKEN",
				URI:           "https://api.tactycal.com",
				Labels:        []string{},
				StatePath:     "/default/path",
				ClientTimeout: time.Second,
			},
		},
		{
			title: "full config",
			data: []byte(`token=TOKEN
			              uri=API_URL/
			              labels=$PATH,label
			              timeout=10s
			              state=/path/to/state`),
			expected: &config{
				Token:         "TOKEN",
				URI:           "API_URL",
				Labels:        []string{os.Getenv("PATH"), "label"},
				ClientTimeout: time.Second * 10,
				StatePath:     "/path/to/state",
			},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.title, func(t *testing.T) {
			s := stubutils.NewStubs(t, &stubutils.ReadFileStub{Path: "path", Output: fixture.data})
			defer s.Close()

			// parse
			c, err := newConfig("path", "/default/path", time.Second)

			if err != nil {
				t.Error("Error should be nil; got", err)
			}

			if !reflect.DeepEqual(fixture.expected, c) {
				t.Errorf("Expected\n%+v\nto equal\n%+v", fixture.expected, c)
			}
		})
	}
}

func TestNewConfig_Errors(t *testing.T) {
	fixtures := []struct {
		title    string
		data     []byte
		expected string
	}{
		{
			title:    "file missing",
			data:     []byte("missing"),
			expected: "configuration: oh no",
		},
		{
			title:    "no token",
			data:     []byte(``),
			expected: "configuration: No token provided",
		},
		{
			title: "broken proxy URL",
			data: []byte(`token=TOKEN
				          proxy=%gh`),
			expected: "configuration: unable to parse proxy URL",
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.title, func(t *testing.T) {
			// // create a temp file
			s := stubutils.NewStubs(t)
			if bytes.Equal(fixture.data, []byte("missing")) {
				s.Add(&stubutils.ReadFileStub{Err: stubutils.ErrOhNo})
			} else {
				s.Add(&stubutils.ReadFileStub{Output: fixture.data})
			}

			// parse
			c, err := newConfig("path", "/default/path", time.Second)

			if c != nil {
				t.Errorf("Expected Config to be nil; got %+v", c)
			}

			if err == nil {
				t.Error("Expected err to not be nil")
			}

			if err != nil && !strings.Contains(err.Error(), fixture.expected) {
				t.Errorf("Expected error to contain \"%s\"; got \"%s\"", fixture.expected, err.Error())
			}
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	parsed, _ := url.Parse(rawURL)
	return parsed
}
