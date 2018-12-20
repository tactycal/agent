package main

import (
	"flag"
	"fmt"

	"github.com/tactycal/agent/packagelookup"
)

func main() {
	configFile := flag.String("f", defaultConfigurationFile, "use a configuration file")
	showVersion := flag.Bool("v", false, "print version and exit")
	debugMode := flag.Bool("d", false, "show debug messages")
	statePath := flag.String("s", defaultStatePath, "path to where tactycal can write its state")
	clientTimeout := flag.Duration("t", defaultClientTimeout, "client timeout for request in seconds")
	localOutput := flag.Bool("l", false, "print host information and installed packages to standard output as json string and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("tactycal %s-%s\n", version, gitCommit)
		return
	}

	if *localOutput {
		if s, err := local(); err != nil {
			fmt.Printf("Failed to retrieved host information and installed packages; %s", err.Error())
		} else {
			fmt.Println(s)
		}

		return
	}

	initLogging(*debugMode)

	log.Infof("Starting tactycal %s-%s", version, gitCommit)

	config, err := newConfig(*configFile, *statePath, *clientTimeout)
	if err != nil {
		log.Fatalf("Failed to read configuration; err = %v", err)
	}

	host, err := getHostInfo()
	if err != nil {
		log.Fatalf("Failed to get host info; err = %v", err)
	}

	packages, err := packagelookup.Get(host.Distribution)
	if err != nil {
		log.Fatalf("Failed to fetch a list of installed packages; err = %v", err)
	}
	log.Debugf("Found %d installed packages", len(packages))

	client := newClient(config, host, newState(*statePath), *clientTimeout)
	if err := client.SendPackageList(packages); err != nil {
		log.Fatalf("Failed to submit list of installed packages; err = %v", err)
	}
}
