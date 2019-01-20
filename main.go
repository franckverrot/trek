package main

import "log"

func main() {
	options := parseFlags()

	switch options.trekMode {
	case NcursesMode:
		showUI(options)
	case OneOffMode:
		showCLI(options)
	case HelpMode:
		usage()
	default:
		log.Panicf("trek: unknown mode %+v\n", options.trekMode)
	}
}
