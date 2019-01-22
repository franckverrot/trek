package main

import (
	"fmt"
	"log"
	"os"
)

func runCommand(trekOptions trekOptions) error {
	trekState := new(trekStateType)

	trekState.nomadConnectConfiguration.addEnvironment("default", trekOptions.nomadAddress)
	trekState.selectedClusterIndex = 0

	if err := trekState.Connect(); err != nil {
		log.Panicln(err)
		return err
	}

	switch trekOptions.trekMode {
	case ListJobsMode:

		if trekOptions.displayFormat == "" {
			trekOptions.displayFormat = jobsListFormat
		}

		provider := jobsFormatProvider{
			Jobs: buildJobs(trekState.Jobs()),
		}
		trekPrintDetails(os.Stdout, trekOptions.displayFormat, provider)

	case JobMode:

		for index, job := range trekState.Jobs() {
			if *job.Name == trekOptions.jobID {
				trekState.selectedJob = index
			}
		}

		if trekOptions.taskGroup == "" {

			if trekOptions.displayFormat == "" {
				trekOptions.displayFormat = taskGroupsListFormat
			}
			provider := jobFormatProvider{
				TaskGroups: buildTaskGroups(trekState.CurrentTaskGroups()),
			}
			trekPrintDetails(os.Stdout, trekOptions.displayFormat, provider)

		} else {
			// Find task group provided by the user
			trekState.selectedAllocationGroup = -1
			for index, tg := range trekState.CurrentTaskGroups() {
				if *tg.Name == trekOptions.taskGroup {
					trekState.selectedAllocationGroup = index
				}
			}

			if trekState.selectedAllocationGroup < 0 || trekState.selectedAllocationGroup > len(trekState.CurrentTaskGroups())-1 {

				// No such task group found, display available ones
				fmt.Printf("Unknown task group.  Available task groups:\n")
				for _, tg := range trekState.CurrentTaskGroups() {
					fmt.Printf("* %s\n", *tg.Name)
				}

			} else {
				// No allocation provided by the user, display all of them
				if trekOptions.allocationIndex == -1 {

					if trekOptions.displayFormat == "" {
						trekOptions.displayFormat = allocationsFormat
					}

					provider := taskGroupFormatProvider{
						Allocations: buildAllocations(trekState.CurrentAllocations()),
					}
					trekPrintDetails(os.Stdout, trekOptions.displayFormat, provider)

				} else {
					trekState.selectedAllocationIndex = trekOptions.allocationIndex

					if trekOptions.allocationIndex < 0 || trekOptions.allocationIndex > len(trekState.CurrentAllocations())-1 {

						// out of bounds, show existing ones
						fmt.Printf("Allocation index %d out-of-bounds.  Valid indices:\n", trekOptions.allocationIndex)
						for index, alloc := range trekState.CurrentAllocations() {
							fmt.Printf("(%d) %s\n", index, alloc.Name)
						}

					} else {

						// Allocation found, no task provided by user
						if trekOptions.taskName == "" {

							if trekOptions.displayFormat == "" {
								trekOptions.displayFormat = allocationDetailsFormat
							}

							provider := allocationFormatProvider{
								IP:    trekState.CurrentAllocation().IP(),
								Tasks: buildTasks(trekState.Tasks()),
							}
							trekPrintDetails(os.Stdout, trekOptions.displayFormat, provider)

						} else {

							// Find task
							trekState.selectedTask = -1
							for index, task := range trekState.Tasks() {
								if task.Name == trekOptions.taskName {
									trekState.selectedTask = index
								}
							}

							// No task? Show all of them
							if trekState.selectedTask == -1 {
								fmt.Printf("Task %s not found.  Available tasks:\n", trekOptions.taskName)
								for _, task := range trekState.Tasks() {
									fmt.Printf("* %s\n", task.Name)
								}
							} else {

								if trekOptions.displayFormat == "" {
									trekOptions.displayFormat = taskDetailsFormat
								}

								alloc := trekState.CurrentAllocation()
								task := trekState.CurrentTask()

								provider := taskFormatProvider{
									Task:        trekTask{Name: task.Name, Driver: task.Driver, Config: task.Config},
									Node:        trekNode{Name: alloc.node.Name, IP: alloc.IP()},
									Network:     buildNetwork(alloc.allocation.TaskResources[task.Name].Networks),
									Environment: buildEnv(task.Env),
								}
								trekPrintDetails(os.Stdout, trekOptions.displayFormat, provider)
							}
						}
					}
				}
			}
		}

	default:
		fmt.Printf("Unknown mode: %s.  Exiting.\n", trekOptions.trekMode)
		os.Exit(1)
	}
	return nil
}
