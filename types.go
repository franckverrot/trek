package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"sort"

	"github.com/hashicorp/nomad/api"
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

func (trekState *trekStateType) getNodeFromAllocation(alloc nomad.Allocation) api.Node {
	options := &nomad.QueryOptions{}
	nodes := trekState.client.Nodes()
	node, _, err := nodes.Info(alloc.NodeID, options)

	if err != nil {
		log.Panicln(err)
	}
	return *node
}

func (trekState *trekStateType) CurrentAllocation() allocation {
	index := trekState.selectedAllocationIndex
	alloc := trekState.foundAllocations[index]
	node := trekState.getNodeFromAllocation(alloc)
	return allocation{allocation: alloc, node: node}
}
func (trekState *trekStateType) CurrentJob() nomad.Job {
	return trekState.jobs[trekState.selectedJob]
}
func (trekState *trekStateType) CurrentTaskGroups() []*nomad.TaskGroup {
	return trekState.CurrentJob().TaskGroups
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
	jobListStubs, _, err := trekState.client.Jobs().List(options)

	if err != nil {
		log.Panicln(err)
	}

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

type taskFormatProvider struct {
	IP          string
	Network     trekCommandNetwork
	Environment trekCommandEnvironment
}

type trekCommandNetwork struct {
	Ports map[string]trekCommandPort
}

type trekCommandPort struct {
	Value    int
	Reserved bool // if not reserved? it's dynamic
}

type trekCommandEnvironment map[string]trekCommandEnvironmentVariable

type trekCommandEnvironmentVariable struct {
	Value string
}

func trekPrintDetails(w io.Writer, format string, data interface{}) {

	tmpl, err := template.
		New("output").
		Funcs(template.FuncMap{
			"Debug":    func(structure interface{}) string { return fmt.Sprintf("DEBUG: %+v\n", structure) },
			"DebugAll": func() string { return fmt.Sprintf("DEBUG ALL: %+v\n", data) },
		}).
		Parse(format)

	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}

type allocationFormatProvider struct {
	IP    string
	Tasks []trekTask
}

type trekTask struct {
	Name string
}

type taskGroupFormatProvider struct {
	Allocations []trekAllocation
}

type trekAllocation struct {
	Name string
}
