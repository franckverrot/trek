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
	selectedCluster           int
	selectedJob               int
	selectedTaskGroup         int
	selectedTask              int
	selectedService           int
	showUI                    bool
	client                    *nomad.Client
	jobs                      []nomad.Job
	nomadConnectConfiguration configuration
}

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

func createView(g *gocui.Gui, trekState *trekStateType, view trekView) error {
	maxX, maxY := g.Size()
	bounds := getBounds(maxX, maxY, view.panelNum, view.panelsTotal, view.margin)
	if v, err := g.SetView(view.name, bounds.startX, bounds.startY, bounds.endX, bounds.endY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = view.name
		view.handler(v, trekState)
	}

	if view.foregroundAfterCreation {
		if _, err := g.SetCurrentView(view.name); err != nil {
			log.Panicln(err)
			return err
		}
	}

	return nil
}

func selectCluster(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return createView(g, trekState,
		trekView{
			name: "Jobs",
			foregroundAfterCreation: true,
			panelNum:                1,
			panelsTotal:             5,
			margin:                  0,
			handler: func(view *gocui.View, trekState *trekStateType) error {
				view.Highlight = true
				view.SelBgColor = gocui.ColorGreen
				view.SelFgColor = gocui.ColorBlack
				view.Editable = false
				view.Wrap = false

				options := &nomad.QueryOptions{}
				jobListStubs, _, _ := trekState.client.Jobs().List(options)
				trekState.jobs = make([]nomad.Job, 0)
				for _, job := range jobListStubs {
					fullJob, _, _ := trekState.client.Jobs().Info(job.ID, options)
					trekState.jobs = append(trekState.jobs, *fullJob)
					fmt.Fprintf(view, "%s (%s)\n", *(fullJob.Name), *(fullJob.ID))
				}

				return nil
			}})
}

func selectJob(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	if len(trekState.jobs) < 1 {
		return nil
	}

	return createView(g, trekState,
		trekView{
			name: "Task Groups",
			foregroundAfterCreation: true,
			panelNum:                2,
			panelsTotal:             5,
			margin:                  0,
			handler: func(view *gocui.View, trekState *trekStateType) error {
				view.Highlight = true
				view.SelBgColor = gocui.ColorGreen
				view.SelFgColor = gocui.ColorBlack
				view.Editable = false
				view.Wrap = false

				job := trekState.jobs[trekState.selectedJob]

				for _, taskGroup := range job.TaskGroups {
					fmt.Fprintf(view, "%s (%d)\n", *(taskGroup.Name), *(taskGroup.Count))
				}

				return nil
			}})
}

func selectTaskGroup(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return createView(g, trekState,
		trekView{
			name: "Tasks",
			foregroundAfterCreation: true,
			panelNum:                3,
			panelsTotal:             5,
			margin:                  0,
			handler: func(view *gocui.View, trekState *trekStateType) error {
				view.Highlight = true
				view.SelBgColor = gocui.ColorGreen
				view.SelFgColor = gocui.ColorBlack
				view.Editable = false
				view.Wrap = false

				taskGroup := trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup]

				for _, task := range taskGroup.Tasks {
					fmt.Fprintf(view, "%s (%s)\n", (task.Name), (task.Driver))
				}

				return nil
			}})
}

func selectTask(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return createView(g, trekState,
		trekView{
			name: "Services",
			foregroundAfterCreation: true,
			panelNum:                4,
			panelsTotal:             5,
			margin:                  0,
			handler: func(view *gocui.View, trekState *trekStateType) error {
				view.Highlight = true
				view.SelBgColor = gocui.ColorGreen
				view.SelFgColor = gocui.ColorBlack
				view.Editable = false
				view.Wrap = false

				task := trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup].Tasks[trekState.selectedTask]

				for _, service := range task.Services {
					fmt.Fprintf(view, "%s\n", (service.Name))
				}

				return nil
			}})
}

func selectService(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return createView(g, trekState,
		trekView{
			name: "Service",
			foregroundAfterCreation: true,
			panelNum:                0,
			panelsTotal:             1,
			margin:                  10,
			handler: func(view *gocui.View, trekState *trekStateType) error {
				view.Highlight = true
				view.SelBgColor = gocui.ColorGreen
				view.SelFgColor = gocui.ColorBlack
				view.Editable = false
				view.Wrap = false

				service := trekState.jobs[trekState.selectedJob].TaskGroups[trekState.selectedTaskGroup].Tasks[trekState.selectedTask].Services[trekState.selectedService]

				val := reflect.Indirect(reflect.ValueOf(service))
				valType := val.Type()

				for i := 0; i < val.NumField(); i++ {
					field := valType.Field(i)
					value := val.FieldByName(field.Name).Interface()
					name := field.Name

					fmt.Fprintf(view, "%s: %+v\n", name, value)
				}

				return nil
			}})
}

func deleteView(currentView string, newCurrentView string, handler deleteViewCallback) uiHandlerWithStateType {
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
		handler: deleteView("Jobs", "Clusters", func(trekState *trekStateType) { trekState.selectedJob = 0 })},
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
		handler: deleteView("Task Groups", "Jobs", func(trekState *trekStateType) { trekState.selectedTaskGroup = 0 })},
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
		handler: deleteView("Tasks", "Task Groups", func(trekState *trekStateType) { trekState.selectedTask = 0 })},
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
		handler: deleteView("Services", "Tasks", func(trekState *trekStateType) { trekState.selectedService = 0 })},
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
		handler: deleteView("Service", "Services", func(trekState *trekStateType) {})},

	binding{panelName: "", key: gocui.KeyCtrlC, handler: quit},
	binding{panelName: "", key: gocui.KeyF12, handler: quit},
	binding{panelName: "msg", key: gocui.KeyEnter,
		handler: deleteView("msg", "Tasks", func(trekState *trekStateType) {})},
}

func keybindings(g *gocui.Gui, trekState *trekStateType) error {
	for _, binding := range bindings {
		if err := g.SetKeybinding(binding.panelName, binding.key, gocui.ModNone, stateify(binding.handler, trekState)); err != nil {
			return err
		}
	}

	return nil
}

type boundsType struct {
	startX int
	startY int
	endX   int
	endY   int
}

func getBounds(maxX int, maxY int, currentPanel int, totalPanels int, margin int) boundsType {
	var startX int
	var width int
	var endX int
	var endY int
	if maxX <= 80 {
		startX = currentPanel * 0
		width = maxX - 1
		endX = width
	} else {
		width = (maxX / totalPanels)
		startX = currentPanel * width
		endX = startX + width - 1
	}
	endY = maxY - 1 // to see the border
	startY := 2     // menu_view

	return boundsType{
		startX: startX + margin,
		startY: startY + margin,
		endX:   endX - margin,
		endY:   endY - margin}
}

var initialized bool = false

func layout(trekState *trekStateType) layoutType {
	return func(g *gocui.Gui) error {
		if initialized {
			return nil
		}
		title := "Trek"

		// Show menu
		maxX, _ := g.Size()
		startX := -1 // no frame
		startY := -1 // no frame
		endX := maxX - 1
		endY := 1
		offset := 0
		if v, err := g.SetView("title_view", startX, startY, endX, endY); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Highlight = true
			v.Frame = false
			v.SelBgColor = gocui.ColorBlue
			v.SelFgColor = gocui.ColorBlack
			fmt.Fprintf(v, "%s", title)
			offset += len(title)
		}

		offset += 2
		menu_items := []string{"F1:DEBUG", "F12:EXIT"}

		if v, err := g.SetView("menu_items", startX+offset, startY, endX, endY); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Highlight = false
			v.Frame = false
			v.BgColor = gocui.ColorGreen
			v.FgColor = gocui.ColorBlack

			for index, optionName := range menu_items {
				if index > 0 {
					fmt.Fprintf(v, " | ")
				}
				fmt.Fprintf(v, "%s", optionName)
			}
		}

		initialized = true
		return createView(g, trekState,
			trekView{
				name: "Clusters",
				foregroundAfterCreation: true,
				panelNum:                0,
				panelsTotal:             5,
				margin:                  0,
				handler: func(view *gocui.View, trekState *trekStateType) error {

					view.Highlight = true
					view.SelBgColor = gocui.ColorGreen
					view.SelFgColor = gocui.ColorBlack
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
						fmt.Fprintf(view, "%s\n", env.Name)
					}

					return nil
				}})
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

func showUI(trekState *trekStateType) {
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
}

func showCLI(trekState *trekStateType) {
	options := &nomad.QueryOptions{}
	allocs := trekState.client.Allocations()
	allocsListStub, _, _ := allocs.List(options)
	foundAllocations := make([]nomad.Allocation, 0)
	for _, stub := range allocsListStub {
		alloc, _, err := allocs.Info(stub.ID, options)
		if err != nil {
			log.Panicln(err)
		}
		if alloc.JobID == jobID {
			foundAllocations = append(foundAllocations, *alloc)
		}
	}
	if len(foundAllocations) == 0 {
		jobsHandle := trekState.client.Jobs()
		jobs, _, _ := jobsHandle.List(nil)

		fmt.Printf("\"%s\" Not found.  Available jobs:\n", jobID)
		for index, job := range jobs {
			fmt.Printf("\t%d) %s\n", index+1, job.ID)
		}
	} else {
		nodes := trekState.client.Nodes()

		for _, foundAllocation := range foundAllocations {
			node, _, err := nodes.Info(foundAllocation.NodeID, options)
			ip := node.Attributes["unique.network.ip-address"]
			if err != nil {
				log.Panicln(err)
			}
			fmt.Printf("%s (%s)\n", foundAllocation.Name, foundAllocation.ID)
			for _, task := range foundAllocation.TaskResources {
				for _, network := range task.Networks {
					for _, dynPort := range network.DynamicPorts {
						fmt.Printf("\t%s => %s:%d\n", dynPort.Label, ip, dynPort.Value)
					}
				}
			}
		}

	}
}

func main() {
	//connect to nomad
	trekState := new(trekStateType)

	var err error
	trekState.client, err = nomad.NewClient(nomad.DefaultConfig())

	parseFlags(trekState)

	if err != nil {
		log.Panicln(err)
	}

	if trekState.showUI {
		showUI(trekState)
	} else {
		showCLI(trekState)
	}
}
