package main

import "github.com/tactycal/agent/osdiscovery"

// Host holds information about a single host
type Host struct {
	Fqdn         string   `json:"fqdn"`
	Distribution string   `json:"distribution"`
	Release      string   `json:"release"`
	Architecture string   `json:"architecture"`
	Kernel       string   `json:"kernel"`
	Labels       []string `json:"labels"`
}

func getHostInfo() (*Host, error) {

	osInfo, err := osdiscovery.Get()
	if err != nil {
		return nil, err
	}

	return &Host{
		Fqdn:         osInfo.Fqdn,
		Distribution: osInfo.Distribution,
		Release:      osInfo.Release,
		Architecture: osInfo.Architecture,
		Kernel:       osInfo.Kernel,
	}, nil
}
