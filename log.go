package main

import (
	"github.com/op/go-logging"
)

var log *logging.Logger

func initLogging(verbose bool) {
	module := ""
	log = logging.MustGetLogger(module)

	if verbose {
		logging.SetLevel(logging.DEBUG, module)
	} else {
		logging.SetLevel(logging.ERROR, module)
	}
}
