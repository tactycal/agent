package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

	packages, err := GetPackages()
	if err != nil {
		log.Fatalf("Failed to fetch a list of installed packages; err = %v", err)
	}
	log.Debugf("Found %d installed packages", len(packages))

	client := NewClient(config, GetHostInfo(), NewState(*statePath), *clientTimeout)
	if err := client.SendPackageList(packages); err != nil {
		log.Fatalf("Failed to submit list of installed packages; err = %v", err)
	}
}

var readFile = func(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

var writeFile = func(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

var execCommand = func(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}
