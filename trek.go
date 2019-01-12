package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/jroimartin/gocui"
)

type configuration struct {
	Environments []environment
}
type environment struct {
	Name    string
	Address string
}
type cursorPosition struct {
	x int
	y int
}
type layoutType func(g *gocui.Gui) error
type clearViewCallback func(trekState *trekStateType)
type cursorCallback func(trekState *trekStateType, position cursorPosition)
type numElementsComputerCallback func(trekState *trekStateType) int
type uiHandlerType func(g *gocui.Gui, v *gocui.View) error
type uiHandlerWithStateType func(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error

type trekStateType struct {
	selectedCluster           int
	selectedJob               int
	selectedTaskGroup         int
	selectedTask              int
	selectedService           int
	showUI                    bool
	client                    *nomad.Client
	jobs                      []nomad.Job
	jobsHandle                *nomad.Jobs
	nomadConnectConfiguration configuration
}

var options *nomad.QueryOptions

// used in CLI mode
var jobID string

func stateify(handler uiHandlerWithStateType, trekState *trekStateType) uiHandlerType {
	return func(g *gocui.Gui, v *gocui.View) error {
		return handler(g, v, trekState)
	}
}

func cursorDown(handler cursorCallback, numElementsComputer numElementsComputerCallback) uiHandlerWithStateType {
	return func(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
		if v != nil {
			cx, cy := v.Cursor()

			if cy >= numElementsComputer(trekState)-1 {
				return nil
			}

			if err := v.SetCursor(cx, cy+1); err != nil {
				ox, oy := v.Origin()
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			} else {
				handler(trekState, cursorPosition{x: cx, y: cy + 1})
			}

		}
		return nil
	}
}

func cursorUp(handler cursorCallback) uiHandlerWithStateType {
	return func(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
		if v != nil {
			ox, oy := v.Origin()
			cx, cy := v.Cursor()
			if cy <= 0 {
				return nil
			}
			if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
				if err := v.SetOrigin(ox, oy-1); err != nil {
					return err
				}
			} else {
				handler(trekState, cursorPosition{x: cx, y: cy - 1})
			}
		}
		return nil
	}
}

func confirmTaskSelection(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, l)
		if _, err := g.SetCurrentView("msg"); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return gocui.ErrQuit
}

func selectCluster(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	_, maxY := g.Size()
	if v, err := g.SetView("Jobs", 30, 2, 60, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Jobs"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		trekState.jobsHandle = trekState.client.Jobs()
		jobListStubs, _, _ := trekState.jobsHandle.List(options)
		trekState.jobs = make([]nomad.Job, 0)
		for _, job := range jobListStubs {
			fullJob, _, _ := trekState.jobsHandle.Info(job.ID, options)
			trekState.jobs = append(trekState.jobs, *fullJob)
			fmt.Fprintf(v, "%s (%s)\n", *(fullJob.Name), *(fullJob.ID))
		}
		v.Editable = false
		v.Wrap = false
	}
	if _, err := g.SetCurrentView("Jobs"); err != nil {
		return err
	}
	return nil
}

func selectJob(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	_, maxY := g.Size()

	if len(trekState.jobs) < 1 {
		return nil
	}

	job := trekState.jobs[trekState.selectedJob]

	if v, err := g.SetView("Task Groups", 60, 2, 90, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Task Groups"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		for _, taskGroup := range job.TaskGroups {
			fmt.Fprintf(v, "%s (%d)\n", *(taskGroup.Name), *(taskGroup.Count))
		}
		v.Editable = false
		v.Wrap = false
	}
	if _, err := g.SetCurrentView("Task Groups"); err != nil {
		return err
	}
	return nil
}
func selectTaskGroup(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	_, maxY := g.Size()
	if v, err := g.SetView("Tasks", 90, 2, 120, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Tasks"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		taskGroup := trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup]

		for _, task := range taskGroup.Tasks {
			fmt.Fprintf(v, "%s (%s)\n", (task.Name), (task.Driver))
		}

		v.Editable = false
		v.Wrap = false
	}
	if _, err := g.SetCurrentView("Tasks"); err != nil {
		return err
	}
	return nil
}

func selectTask(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	// confirmTaskSelection(g, v)
	_, maxY := g.Size()
	if v, err := g.SetView("Services", 120, 2, 150, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Services"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		task := trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup].Tasks[trekState.selectedTask]

		for _, service := range task.Services {
			fmt.Fprintf(v, "%s\n", (service.Name))
		}

		v.Editable = false
		v.Wrap = false
	}
	if _, err := g.SetCurrentView("Services"); err != nil {
		return err
	}
	return nil
}

func selectService(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("Service", 20, 20, maxX-20, maxY-20); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Service"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		service := trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup].Tasks[trekState.selectedTask].Services[trekState.selectedService]

		val := reflect.Indirect(reflect.ValueOf(service))
		valType := val.Type()

		for i := 0; i < val.NumField(); i++ {
			field := valType.Field(i)
			value := val.FieldByName(field.Name).Interface()
			name := field.Name

			fmt.Fprintf(v, "%s: %+v\n", name, value)
		}

		v.Editable = false
		v.Wrap = false
	}
	if _, err := g.SetCurrentView("Service"); err != nil {
		return err
	}
	return nil
}

func clearView(currentView string, newCurrentView string, handler clearViewCallback) uiHandlerWithStateType {
	return func(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
		if err := g.DeleteView(currentView); err != nil {
			return err
		}
		if _, err := g.SetCurrentView(newCurrentView); err != nil {
			return err
		}
		handler(trekState)
		return nil
	}
}

// binding is some binding
type binding struct {
	panelName string
	key       gocui.Key
	handler   uiHandlerWithStateType
}

var bindings = []binding{
	binding{panelName: "Clusters", key: gocui.KeyEnter, handler: selectCluster},
	binding{panelName: "Clusters", key: gocui.KeyArrowRight, handler: selectCluster},
	binding{panelName: "Clusters", key: gocui.KeyArrowDown, handler: cursorDown(
		func(trekState *trekStateType, position cursorPosition) { trekState.selectedCluster = position.y },
		func(trekState *trekStateType) int { return len(trekState.nomadConnectConfiguration.Environments) })},
	binding{panelName: "Clusters", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedCluster = position.y
		})},

	binding{panelName: "Jobs", key: gocui.KeyArrowLeft,
		handler: clearView("Jobs", "Clusters", func(trekState *trekStateType) { trekState.selectedJob = 0 })},
	binding{panelName: "Jobs", key: gocui.KeyEnter, handler: selectJob},
	binding{panelName: "Jobs", key: gocui.KeyArrowRight, handler: selectJob},
	binding{panelName: "Jobs", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedJob = position.y
		})},
	binding{panelName: "Jobs", key: gocui.KeyArrowDown, handler: cursorDown(
		func(trekState *trekStateType, position cursorPosition) { trekState.selectedJob = position.y },
		func(trekState *trekStateType) int { return len(trekState.jobs) })},

	binding{panelName: "Task Groups", key: gocui.KeyArrowLeft,
		handler: clearView("Task Groups", "Jobs", func(trekState *trekStateType) { trekState.selectedTaskGroup = 0 })},
	binding{panelName: "Task Groups", key: gocui.KeyEnter, handler: selectTaskGroup},
	binding{panelName: "Task Groups", key: gocui.KeyArrowRight, handler: selectTaskGroup},
	binding{panelName: "Task Groups", key: gocui.KeyArrowDown, handler: cursorDown(
		func(trekState *trekStateType, position cursorPosition) { trekState.selectedTaskGroup = position.y },
		func(trekState *trekStateType) int { return len(trekState.jobs[trekState.selectedJob].TaskGroups) })},
	binding{panelName: "Task Groups", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedTaskGroup = position.y
		})},

	binding{panelName: "Tasks", key: gocui.KeyArrowLeft,
		handler: clearView("Tasks", "Task Groups", func(trekState *trekStateType) { trekState.selectedTask = 0 })},
	binding{panelName: "Tasks", key: gocui.KeyEnter, handler: selectTask},
	binding{panelName: "Tasks", key: gocui.KeyArrowRight, handler: selectTask},
	binding{panelName: "Tasks", key: gocui.KeyArrowDown,
		handler: cursorDown(
			func(trekState *trekStateType, position cursorPosition) { trekState.selectedTask = position.y },
			func(trekState *trekStateType) int {
				return len(trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup].Tasks)
			})},
	binding{panelName: "Tasks", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedTask = position.y
		})},

	binding{panelName: "Services", key: gocui.KeyArrowLeft,
		handler: clearView("Services", "Tasks", func(trekState *trekStateType) { trekState.selectedService = 0 })},
	binding{panelName: "Services", key: gocui.KeyEnter, handler: selectService},
	binding{panelName: "Services", key: gocui.KeyArrowRight, handler: selectService},
	binding{panelName: "Services", key: gocui.KeyArrowDown,
		handler: cursorDown(
			func(trekState *trekStateType, position cursorPosition) { trekState.selectedService = position.y },
			func(trekState *trekStateType) int {
				return len(trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup].Tasks[trekState.selectedTask].Services)
			})},
	binding{panelName: "Services", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedService = position.y
		})},

	binding{panelName: "Service", key: gocui.KeyEnter,
		handler: clearView("Service", "Services", func(trekState *trekStateType) {})},

	binding{panelName: "", key: gocui.KeyCtrlC, handler: quit},
	binding{panelName: "msg", key: gocui.KeyEnter,
		handler: clearView("msg", "Tasks", func(trekState *trekStateType) {})},
}

func keybindings(g *gocui.Gui, trekState *trekStateType) error {
	for _, binding := range bindings {
		if err := g.SetKeybinding(binding.panelName, binding.key, gocui.ModNone, stateify(binding.handler, trekState)); err != nil {
			return err
		}
	}

	return nil
}

func layout(trekState *trekStateType) layoutType {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		title := "Welcome to Nomad Connect!"
		padding := (maxX-1)/2 - (len(title) / 2)
		if v, err := g.SetView("title_padding", padding, 0, padding+len(title)+1, 2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Highlight = true
			v.Frame = false
			v.SelBgColor = gocui.ColorBlue
			v.SelFgColor = gocui.ColorBlack
			fmt.Fprintf(v, "%*s", 5, title)
		}
		if v, err := g.SetView("Clusters", 0, 2, 30, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
			v.Title = "Clusters"
			file, err := os.Open(".trek.rc")
			if err != nil {
				log.Panicln(err)
			}
			decoder := json.NewDecoder(file)
			trekState.nomadConnectConfiguration = configuration{}
			err = decoder.Decode(&trekState.nomadConnectConfiguration)
			if err != nil {
				log.Panicln(err)
			}

			for _, env := range trekState.nomadConnectConfiguration.Environments {
				fmt.Fprintf(v, "%s\n", env.Name)
			}

			if _, err := g.SetCurrentView("Clusters"); err != nil {
				return err
			}
		}
		return nil
	}
}

func checkValidFlag(flagName string, flagValue string, validValues map[string]bool) {
	if !validValues[flagValue] {
		usage()

		keys := reflect.ValueOf(validValues).MapKeys()
		strkeys := make([]string, len(keys))
		for i := 0; i < len(keys); i++ {
			strkeys[i] = keys[i].String()
		}
		fmt.Fprintf(os.Stderr, "\nbad value for %s, got %s, accepting: %s\n", flagName, flagValue, strings.Join(strkeys, ", "))
		os.Exit(1)
	}
}
func parseFlags(trekState *trekStateType) {
	flag.BoolVar(&trekState.showUI, "ui", true, "whether to show the ncurses UI or not")
	flag.StringVar(&jobID, "jobID", "", "jobID to get")

	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [inputfile]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	//connect to nomad
	trekState := new(trekStateType)
	options = &nomad.QueryOptions{}

	var err error
	trekState.client, err = nomad.NewClient(nomad.DefaultConfig())

	parseFlags(trekState)

	if err != nil {
		log.Panicln(err)
	}

	if trekState.showUI {
		// build ui
		g, err := gocui.NewGui(gocui.OutputNormal)
		if err != nil {
			log.Panicln(err)
		}
		defer g.Close()

		g.Cursor = true

		g.SetManagerFunc(layout(trekState))

		if err := keybindings(g, trekState); err != nil {
			log.Panicln(err)
		}

		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Panicln(err)
		}
	} else {
		fmt.Printf("Trying to get jobID \"%s\" (specified by -jobID)\n", jobID)
		allocs := trekState.client.Allocations()
		allocsListStub, _, _ := allocs.List(options)
		found := false
		var foundAllocation *nomad.Allocation
		for _, stub := range allocsListStub {
			alloc, _, err := allocs.Info(stub.ID, options)
			if err != nil {
				log.Panicln(err)
			}
			if alloc.JobID == jobID && found == false {
				found = true
				foundAllocation = alloc
			}
		}
		if found == false {
			log.Panicf("Couldn't find the node onto which the job ID %s is running... Aborting\n", jobID)
		} else {
			nodes := trekState.client.Nodes()
			node, _, err := nodes.Info(foundAllocation.NodeID, options)
			if err != nil {
				log.Panicln(err)
			}
			fmt.Printf("%+v\n", node.Attributes["unique.network.ip-address"])
			for _, service := range foundAllocation.Services {
				fmt.Printf("Services %+v\n", service)
			}

			// jobs := client.Jobs()
			// job, _, err := jobs.Info(jobID, options)
		}
	}
}
