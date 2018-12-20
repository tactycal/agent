package main

import (
	"encoding/json"

	"github.com/tactycal/agent/packagelookup"
)

// local returns host information and installed packages as json string
func local() (string, error) {
	host, err := getHostInfo()
	if err != nil {
		return "", err
	}

	packages, err := packagelookup.Get(host.Distribution)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(&sendPackagesRequestBody{host, packages})
	return string(b), err
}
