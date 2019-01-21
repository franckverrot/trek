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

	// JobMode is used for one off commands
	JobMode UIMode = "job"

	// HelpMode is used when we show the help
	HelpMode UIMode = "help"

	// ListJobsMode is used to list jobs
	ListJobsMode UIMode = "list-jobs"
)

type trekOptions struct {
	nomadAddress    string
	trekMode        UIMode
	jobID           string
	taskGroup       string
	allocationIndex int
	taskName        string
	displayFormat   string
}

type cliOptions struct {
	nomadAddress    string
	help            bool
	ncurses         bool
	listJobs        bool
	job             string
	taskGroup       string
	allocationIndex int
	taskName        string
	displayFormat   string
}

func (options *cliOptions) DetermineMode() UIMode {
	actualMode := HelpMode

	if options.help {
		actualMode = HelpMode
	} else {
		if options.ncurses {
			actualMode = NcursesMode
		} else if options.listJobs {
			actualMode = ListJobsMode
		} else if options.job != "" {
			actualMode = JobMode
		}
	}

	return actualMode
}

func parseFlags() trekOptions {
	options := new(cliOptions)
	flag.BoolVar(&(*options).help, "help", false, "show usage prompt")
	flag.StringVar(&(*options).nomadAddress, "nomad-address", "http://localhost:4646", "nomad cluster address")
	flag.BoolVar(&(*options).ncurses, "ui", false, "use UI mode")
	flag.BoolVar(&(*options).listJobs, "list-jobs", false, "list jobs")
	flag.StringVar(&(*options).job, "job", "", "job name to get (only used when running in non-ui mode)")
	flag.StringVar(&(*options).taskGroup, "task-group", "", "task group to get (only used when running in non-ui mode)")
	flag.IntVar(&(*options).allocationIndex, "allocation", -1, "allocation index to get (starts at 0, only used when running in non-ui mode)")
	flag.StringVar(&(*options).taskName, "task", "", "task name to get (only used when running in non-ui mode)")
	flag.StringVar(&(*options).displayFormat, "display-format", "", "task display format")

	flag.Parse()

	return trekOptions{
		nomadAddress:    (*options).nomadAddress,
		trekMode:        (*options).DetermineMode(),
		jobID:           (*options).job,
		taskGroup:       (*options).taskGroup,
		allocationIndex: (*options).allocationIndex,
		taskName:        (*options).taskName,
		displayFormat:   (*options).displayFormat,
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [options]\n", os.Args[0])
	flag.PrintDefaults()
}
