package main

import (
	"log"
	"sort"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/jroimartin/gocui"
)

type configuration struct {
	Environments *[]environment
}

type environment struct {
	Name    string
	Address string
}

func (config *configuration) addEnvironment(name string, address string) {
	if config.Environments == nil {
		config.Environments = new([]environment)
	}
	*config.Environments = append(*config.Environments, environment{Name: name, Address: address})
}

type cursorPosition struct {
	x int
	y int
}

type trekView struct {
	name                    string
	foregroundAfterCreation bool
	panelNum                int
	panelsTotal             int
	margin                  int
	handler                 viewHandlerCallback
}

type layoutType func(g *gocui.Gui) error
type viewHandlerCallback func(view *gocui.View, trekState *trekStateType) error
type deleteViewCallback func(trekState *trekStateType)
type cursorCallback func(trekState *trekStateType, position cursorPosition)
type numElementsComputerCallback func(trekState *trekStateType) int
type uiHandlerType func(g *gocui.Gui, v *gocui.View) error
type uiHandlerWithStateType func(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error

type trekStateType struct {
	selectedClusterIndex      int
	selectedJob               int
	selectedAllocationGroup   int
	selectedAllocationIndex   int
	foundAllocations          []nomad.Allocation
	selectedTask              int
	foundTasks                []nomad.Task
	client                    *nomad.Client
	jobs                      []nomad.Job
	nomadConnectConfiguration configuration
}

func (trekState *trekStateType) CurrentEnvironment() environment {
	return (*trekState.nomadConnectConfiguration.Environments)[trekState.selectedClusterIndex]
}

type allocation struct {
	allocation nomad.Allocation
	node       nomad.Node
}

func (alloc allocation) IP() string {
	return alloc.node.Attributes["unique.network.ip-address"]
}

func (trekState *trekStateType) CurrentAllocation() allocation {
	alloc := trekState.foundAllocations[trekState.selectedAllocationIndex]
	options := &nomad.QueryOptions{}
	nodes := trekState.client.Nodes()
	node, _, err := nodes.Info(alloc.NodeID, options)

	if err != nil {
		log.Panicln(err)
	}
	return allocation{allocation: alloc, node: *node}
}
func (trekState *trekStateType) CurrentJob() nomad.Job {
	return trekState.jobs[trekState.selectedJob]
}
func (trekState *trekStateType) CurrentTaskGroup() nomad.TaskGroup {
	return *trekState.CurrentJob().TaskGroups[trekState.selectedAllocationGroup]
}
func (trekState *trekStateType) Tasks() []*nomad.Task {
	return trekState.CurrentTaskGroup().Tasks
}
func (trekState *trekStateType) CurrentTask() *nomad.Task {
	return trekState.Tasks()[trekState.selectedTask]
}

func (trekState *trekStateType) CurrentAllocations() []nomad.Allocation {
	options := &nomad.QueryOptions{}
	allocs := trekState.client.Allocations()
	allocsListStub, _, _ := allocs.List(options)

	trekState.foundAllocations = make([]nomad.Allocation, 0)

	taskGroup := trekState.CurrentTaskGroup()

	for _, stub := range allocsListStub {
		alloc, _, err := allocs.Info(stub.ID, options)
		if err != nil {
			log.Panicln(err)
		}
		if alloc.TaskGroup == *taskGroup.Name {
			if alloc.ClientStatus == "running" {
				trekState.foundAllocations = append(trekState.foundAllocations, *alloc)
			}
		}
	}
	sort.SliceStable(trekState.foundAllocations, func(i, j int) bool { return trekState.foundAllocations[i].Name < trekState.foundAllocations[j].Name })
	return trekState.foundAllocations
}

func (trekState *trekStateType) Jobs() []nomad.Job {
	options := &nomad.QueryOptions{}
	jobListStubs, _, _ := trekState.client.Jobs().List(options)
	trekState.jobs = make([]nomad.Job, 0)
	for _, job := range jobListStubs {
		fullJob, _, _ := trekState.client.Jobs().Info(job.ID, options)
		trekState.jobs = append(trekState.jobs, *fullJob)
	}
	return trekState.jobs
}

func (trekState *trekStateType) Connect() error {
	config := nomad.DefaultConfig()
	config.Address = trekState.CurrentEnvironment().Address
	var err error
	trekState.client, err = nomad.NewClient(config)

	if err != nil {
		return err
	}
	return nil
}

type boundsType struct {
	startX int
	startY int
	endX   int
	endY   int
}

// binding is some binding
type binding struct {
	panelName string
	key       gocui.Key
	handler   uiHandlerWithStateType
}

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
