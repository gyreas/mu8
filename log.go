package main

import (
	"fmt"
	"log"
	"os"
)

const (
	logfile_path = "mu8.log"
	prefix       = "[Mu8] "
	logflags     = log.Lmsgprefix
)

var logfile *os.File = os.Stderr

func init_logger() bool {
	f, err := os.Create(logfile_path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		return false
	}

	logfile = f

	return true
}

func logger() *log.Logger {
	return log.New(logfile, prefix, logflags)
}
