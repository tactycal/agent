package main

import (
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
	flag.Parse()

	if *showVersion {
		fmt.Printf("tactycal %s-%s\n", Version, GitCommit)
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
