package main

import "github.com/tactycal/agent/osDiscovery"

type Host struct {
	Fqdn         string   `json:"fqdn"`
	Distribution string   `json:"distribution"`
	Release      string   `json:"release"`
	Architecture string   `json:"architecture"`
	Kernel       string   `json:"kernel"`
	Labels       []string `json:"labels"`
}

func GetHostInfo() (*Host, error) {

	osInfo, err := osDiscovery.Get()
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
