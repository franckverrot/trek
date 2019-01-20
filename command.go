package main

import (
	"fmt"
	"log"

	nomad "github.com/hashicorp/nomad/api"
)

func runCommand(trekOptions trekOptions) {
	var err error

	trekState := new(trekStateType)
	trekState.client, err = nomad.NewClient(nomad.DefaultConfig())

	if err != nil {
		log.Panicln(err)
	}

	nomadOptions := &nomad.QueryOptions{}
	allocs := trekState.client.Allocations()
	allocsListStub, _, _ := allocs.List(nomadOptions)
	foundAllocations := make([]nomad.Allocation, 0)
	for _, stub := range allocsListStub {
		alloc, _, err := allocs.Info(stub.ID, nomadOptions)
		if err != nil {
			log.Panicln(err)
		}
		if alloc.JobID == trekOptions.jobID {
			foundAllocations = append(foundAllocations, *alloc)
		}
	}
	if len(foundAllocations) == 0 {
		jobsHandle := trekState.client.Jobs()
		jobs, _, _ := jobsHandle.List(nil)

		fmt.Printf("\"%s\" Not found.  Available jobs:\n", trekOptions.jobID)
		for index, job := range jobs {
			fmt.Printf("\t%d) %s\n", index+1, job.ID)
		}
	} else {
		nodes := trekState.client.Nodes()

		for _, foundAllocation := range foundAllocations {
			node, _, err := nodes.Info(foundAllocation.NodeID, nomadOptions)
			ip := node.Attributes["unique.network.ip-address"]
			if err != nil {
				log.Panicln(err)
			}
			fmt.Printf("%s (%s)\n", foundAllocation.Name, foundAllocation.ID)
			for _, taskResource := range foundAllocation.TaskResources {
				for _, network := range taskResource.Networks {
					for _, dynPort := range network.DynamicPorts {
						fmt.Printf("\t%s => %s:%d\n", dynPort.Label, ip, dynPort.Value)
					}
				}
			}
		}
	}
}
