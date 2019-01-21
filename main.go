package main

import "log"

func main() {
	options := parseFlags()

	switch options.trekMode {
	case NcursesMode:
		runUI(options)
	case ListJobsMode, JobMode:
		runCommand(options)
	case HelpMode:
		usage()
	default:
		log.Panicf("trek: unknown mode %+v\n", options.trekMode)
	}
}
