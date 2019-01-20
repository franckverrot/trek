package main

import (
	"flag"
	"fmt"
	"os"
)

// UIMode describes how the app should run
type UIMode string

const (
	// NcursesMode is used when we're using ncurses
	NcursesMode UIMode = "ui"

	// OneOffMode is used for one off commands
	OneOffMode UIMode = "no-ui"

	// HelpMode is used when we show the help
	HelpMode UIMode = "help"
)

type trekOptions struct {
	jobID    string
	trekMode UIMode
}

type cliOptions struct {
	jobID   string
	help    bool
	ncurses bool
}

func (options *cliOptions) DetermineMode() UIMode {
	actualMode := HelpMode

	if options.help {
		actualMode = HelpMode
	} else {
		if options.ncurses {
			actualMode = NcursesMode
		} else if options.jobID != "" {
			actualMode = OneOffMode
		}
	}

	return actualMode
}

func parseFlags() trekOptions {
	options := new(cliOptions)
	flag.BoolVar(&(*options).help, "help", false, "show usage prompt")
	flag.BoolVar(&(*options).ncurses, "ui", false, "use UI mode")
	flag.StringVar(&(*options).jobID, "jobID", "", "jobID to get (only used when running in non-ui mode)")

	flag.Parse()

	return trekOptions{
		jobID:    (*options).jobID,
		trekMode: (*options).DetermineMode(),
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
	flag.PrintDefaults()
}
