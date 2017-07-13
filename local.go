package main

import (
	"encoding/json"

	"github.com/tactycal/agent/packageLookup"
)

// local returns host information and installed packages as json string
func local() (string, error) {
	host, err := GetHostInfo()
	if err != nil {
		return "", err
	}

	packages, err := packageLookup.Get(host.Distribution)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(&SendPackagesRequestBody{host, packages})
	return string(b), err
}
