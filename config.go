package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultConfigurationFile = "/etc/tactycal/agent.conf"
	DefaultApiUri            = "https://api.tactycal.com/v1"
)

type Config struct {
	Token         string
	Uri           string
	Proxy         *url.URL
	Labels        []string
	StatePath     string
	ClientTimeout time.Duration
}

func NewConfig(file, statePath string, clientTimeout time.Duration) (*Config, error) {
	cfg, err := readConfigurationFromFile(file)
	if err != nil {
		return nil, errors.New("configuration: " + err.Error())
	}

	// set defaults
	if cfg["uri"] == "" {
		cfg["uri"] = DefaultApiUri
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
		return nil, errors.New("configuration: No token provided.")
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
	config := &Config{
		Token:         cfg["token"],
		Uri:           cfg["uri"],
		Proxy:         urlProxy,
		Labels:        cleanSplit(cfg["labels"]),
		ClientTimeout: timeout,
		StatePath:     cfg["state"],
	}
	return config, nil
}

func readConfigurationFromFile(file string) (map[string]string, error) {
	data, err := readFile(file)
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
			arr[key] = Getenv(arr[key][1:], arr[key][1:])
		}
	}

	return arr
}

// Retuns value of environment variable `name`. If env. variable is empty the
// defaultValue is returned.
func Getenv(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}
