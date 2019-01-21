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
		for _, job := range trekState.Jobs() {
			fmt.Printf("* %s\n", *job.Name)
		}

	case JobMode:

		for index, job := range trekState.Jobs() {
			if *job.Name == trekOptions.jobID {
				trekState.selectedJob = index
			}
		}

		if trekOptions.taskGroup == "" {
			// If no task group provided by the user, display the available ones
			for _, tg := range trekState.CurrentTaskGroups() {
				fmt.Printf("* %s\n", *tg.Name)
			}
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
				os.Exit(1)
			} else {
				// No allocation provided by the user, display all of them
				if trekOptions.allocationIndex == -1 {
					// Show task group's allocation
					for index, alloc := range trekState.CurrentAllocations() {
						fmt.Printf("(%d) %s\n", index, alloc.Name)
					}
				} else {
					trekState.selectedAllocationIndex = trekOptions.allocationIndex
					if trekOptions.allocationIndex < 0 || trekOptions.allocationIndex > len(trekState.CurrentAllocations())-1 {
						// out of bounds, show existing ones
						fmt.Printf("Allocation index %d out-of-bounds.  Valid indices:\n", trekOptions.allocationIndex)
						for index, alloc := range trekState.CurrentAllocations() {
							fmt.Printf("(%d) %s\n", index, alloc.Name)
						}
					} else {
						//////////////////
						// No task provided by user, display all of them
						if trekOptions.taskName == "" {
							for _, task := range trekState.Tasks() {
								fmt.Printf("* %s\n", task.Name)
							}
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
								// Show task details
								trekState.FprintCurrentTask(os.Stdout)
							}

						}
					}

				}
			}
		}
		// for _, foundAllocation := range foundAllocations {
		// 	node, _, err := nodes.Info(foundAllocation.NodeID, nomadOptions)
		// 	ip := node.Attributes["unique.network.ip-address"]
		// 	if err != nil {
		// 		log.Panicln(err)
		// 	}
		// 	fmt.Printf("%s (%s)\n", foundAllocation.Name, foundAllocation.ID)
		// 	for _, taskResource := range foundAllocation.TaskResources {
		// 		for _, network := range taskResource.Networks {
		// 			for _, dynPort := range network.DynamicPorts {
		// 				fmt.Printf("\t%s => %s:%d\n", dynPort.Label, ip, dynPort.Value)
		// 			}
		// 		}
		// 	}

	default:
		fmt.Printf("Unknown mode: %s.  Exiting.\n", trekOptions.trekMode)
		os.Exit(1)
	}
	return nil
}
