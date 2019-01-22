package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/nomad/api"
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
								if trekOptions.displayFormat == "" {
									trekState.FprintCurrentTask(os.Stdout)
								} else {
									alloc := trekState.CurrentAllocation()
									task := trekState.CurrentTask()

									provider := taskFormatProvider{
										IP:          alloc.IP(),
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
		}

	default:
		fmt.Printf("Unknown mode: %s.  Exiting.\n", trekOptions.trekMode)
		os.Exit(1)
	}
	return nil
}

func buildNetwork(networks []*api.NetworkResource) trekCommandNetwork {
	network := trekCommandNetwork{}
	network.Ports = map[string]trekCommandPort{}

	if len(networks) > 1 {
		log.Panicf("Found more than one network.  Exiting...")
	}
	for _, reservedPort := range networks[0].ReservedPorts {
		network.Ports[reservedPort.Label] = trekCommandPort{Value: reservedPort.Value, Reserved: true}
	}
	for _, dynPort := range networks[0].DynamicPorts {
		network.Ports[dynPort.Label] = trekCommandPort{Value: dynPort.Value, Reserved: false}
	}

	return network
}

func buildEnv(env map[string]string) trekCommandEnvironment {
	resultingEnv := trekCommandEnvironment{}
	for key, value := range env {
		resultingEnv[key] = trekCommandEnvironmentVariable{Value: value}
	}
	return resultingEnv
}
