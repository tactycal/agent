package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tactycal/agent/stubutils"
)

const (
	defaultConfigurationFile = "/etc/tactycal/agent.conf"
	defaultAPIURI            = "https://api.tactycal.com"
)

type config struct {
	Token         string
	URI           string
	Proxy         *url.URL
	Labels        []string
	StatePath     string
	ClientTimeout time.Duration
}

func newConfig(file, statePath string, clientTimeout time.Duration) (*config, error) {
	cfg, err := readConfigurationFromFile(file)
	if err != nil {
		return nil, errors.New("configuration: " + err.Error())
	}

	// set defaults
	if cfg["uri"] == "" {
		cfg["uri"] = defaultAPIURI
	} else {
		// trim trailing slashes
		cfg["uri"] = strings.TrimRight(cfg["uri"], "/")
	}

	// @todo fix proxy handling
	urlProxy, err := url.Parse(cfg["proxy"])
	if err != nil {
		return nil, errors.New("configuration: unable to parse proxy URL, reason:" + err.Error())
	}
	if urlProxy.String() == "" {
		urlProxy = nil
	}

	// check token is set
	if cfg["token"] == "" {
		return nil, errors.New("configuration: No token provided")
	}

	// set client timeout
	timeout := clientTimeout
	if cfg["timeout"] != "" {
		d, err := time.ParseDuration(cfg["timeout"])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse timeout; reason: %s", err.Error())
		}
		timeout = d
	}

	// set state
	if cfg["state"] == "" {
		cfg["state"] = statePath
	}

	// finally! :)
	c := &config{
		Token:         cfg["token"],
		URI:           cfg["uri"],
		Proxy:         urlProxy,
		Labels:        cleanSplit(cfg["labels"]),
		ClientTimeout: timeout,
		StatePath:     cfg["state"],
	}
	return c, nil
}

func readConfigurationFromFile(file string) (map[string]string, error) {
	data, err := stubutils.ReadFile(file)
	if err != nil {
		return nil, err
	}
	matches := make(map[string]string)
	// no major restrictions for now
	// matches are trimmed later on, no extra check are done at this time
	pattern := "(.*)=(.*)"
	m := regexp.MustCompile(pattern).FindAllStringSubmatch(string(data), -1)

	for _, v := range m {
		key := strings.TrimSpace(v[1])
		value := strings.TrimSpace(v[2])
		matches[key] = strings.Trim(value, `"`)
	}
	return matches, nil
}

func cleanSplit(str string) []string {
	if str == "" {
		return []string{}
	}

	// split string by ,
	arr := strings.Split(str, ",")

	// trim all values
	for key, val := range arr {
		arr[key] = strings.TrimSpace(val)

		// check if we need to substitute any of the values
		if arr[key][:1] == `$` {
			arr[key] = getenv(arr[key][1:], arr[key][1:])
		}
	}

	return arr
}

// getenv returns value of environment variable `name`. If env. variable is empty,
// defaultValue is returned.
func getenv(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}
