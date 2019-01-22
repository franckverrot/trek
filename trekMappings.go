package main

import (
	"log"

	"github.com/hashicorp/nomad/api"
	nomad "github.com/hashicorp/nomad/api"
)

func buildNetwork(networks []*api.NetworkResource) trekCommandNetwork {
	network := trekCommandNetwork{}
	network.DynamicPorts = make([]trekCommandPort, 0)
	network.ReservedPorts = make([]trekCommandPort, 0)

	if len(networks) > 1 {
		log.Panicf("Found more than one network.  Exiting...")
	}
	for _, reservedPort := range networks[0].ReservedPorts {
		network.ReservedPorts = append(network.ReservedPorts, trekCommandPort{Name: reservedPort.Label, Number: reservedPort.Value})
	}
	for _, dynPort := range networks[0].DynamicPorts {
		network.DynamicPorts = append(network.DynamicPorts, trekCommandPort{Name: dynPort.Label, Number: dynPort.Value})
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

func buildTasks(tasks []*nomad.Task) []trekTask {
	result := make([]trekTask, 0)

	for _, task := range tasks {
		result = append(result, trekTask{Name: task.Name})
	}

	return result
}

func buildAllocations(allocs []nomad.Allocation) []trekAllocation {
	result := make([]trekAllocation, 0)

	for _, alloc := range allocs {
		result = append(result, trekAllocation{Name: alloc.Name})
	}

	return result
}

func buildTaskGroups(tgs []*nomad.TaskGroup) []trekTaskGroup {
	result := make([]trekTaskGroup, 0)

	for _, taskGroup := range tgs {
		result = append(result, trekTaskGroup{Name: *taskGroup.Name})
	}

	return result
}

func buildJobs(jobs []nomad.Job) []trekJob {
	result := make([]trekJob, 0)

	for _, job := range jobs {
		result = append(result, trekJob{Name: *job.Name})
	}

	return result
}
