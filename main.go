package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/tactycal/agent/packageLookup"
)

func main() {
	configFile := flag.String("f", DefaultConfigurationFile, "use a configuration file")
	showVersion := flag.Bool("v", false, "print version and exit")
	debugMode := flag.Bool("d", false, "show debug messages")
	statePath := flag.String("s", DefaultStatePath, "path to where tactycal can write its state")
	clientTimeout := flag.Duration("t", DefaultClientTimeout, "client timeout for request in seconds")
	localOutput := flag.Bool("l", false, "print host information and installed packages to standard output as json string and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("tactycal %s-%s\n", Version, GitCommit)
		return
	}

	if *localOutput {
		local()
		return
	}

	initLogging(*debugMode)

	log.Infof("Starting tactycal %s-%s", Version, GitCommit)

	config, err := NewConfig(*configFile, *statePath, *clientTimeout)
	if err != nil {
		log.Fatalf("Failed to read configuration; err = %v", err)
	}

	host, err := GetHostInfo()
	if err != nil {
		log.Fatalf("Failed to get host info; err = %v", err)
	}

	packages, err := packageLookup.Get(host.Distribution)
	if err != nil {
		log.Fatalf("Failed to fetch a list of installed packages; err = %v", err)
	}
	log.Debugf("Found %d installed packages", len(packages))

	client := NewClient(config, host, NewState(*statePath), *clientTimeout)
	if err := client.SendPackageList(packages); err != nil {
		log.Fatalf("Failed to submit list of installed packages; err = %v", err)
	}
}

func local() {
	host, err := GetHostInfo()
	if err != nil {
		fmt.Printf("Failed to get host info; err = %v", err)
		return
	}

	packages, err := packageLookup.Get(host.Distribution)
	if err != nil {
		fmt.Printf("Failed to fetch a list of installed packages; err = %v", err)
		return
	}

	b, _ := json.Marshal(&SendPackagesRequestBody{host, packages})
	fmt.Println(string(b))
}
