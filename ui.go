package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

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

func createView(g *gocui.Gui, view trekView, trekState *trekStateType) error {
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
	return createView(g,
		trekView{
			name:                    "Jobs",
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

				if err := trekState.Connect(); err != nil {
					log.Panicln(err)
					return nil
				}

				for _, job := range trekState.Jobs() {
					fmt.Fprintf(view, "%s (%s)\n", *(job.Name), *(job.ID))
				}

				return nil
			},
		},
		trekState,
	)
}

func selectJob(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	if len(trekState.Jobs()) < 1 {
		return nil
	}

	return createView(g,
		trekView{
			name:                    "Task Groups",
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

				job := trekState.CurrentJob()

				for _, taskGroup := range job.TaskGroups {
					fmt.Fprintf(view, "%s (%d)\n", *(taskGroup.Name), *(taskGroup.Count))
				}

				return nil
			},
		},
		trekState,
	)
}

func selectTaskGroup(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return createView(g,
		trekView{
			name:                    "Allocations",
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

				for _, all := range trekState.CurrentAllocations() {
					fmt.Fprintf(view, "%s\n", all.Name)
				}

				return nil
			},
		},
		trekState,
	)
}

func selectAllocation(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return createView(g,
		trekView{
			name:                    "Tasks",
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

				for _, task := range trekState.Tasks() {
					fmt.Fprintf(view, "%s\n", task.Name)
				}

				return nil
			},
		},
		trekState,
	)
}

func selectTask(g *gocui.Gui, v *gocui.View, trekState *trekStateType) error {
	return createView(g,
		trekView{
			name:                    "Task",
			foregroundAfterCreation: true,
			panelNum:                0,
			panelsTotal:             1,
			margin:                  10,
			handler: func(view *gocui.View, trekState *trekStateType) error {
				view.SelBgColor = gocui.ColorGreen
				view.SelFgColor = gocui.ColorBlack
				view.Editable = false
				view.Wrap = false

				task := trekState.CurrentTask()

				currentAllocation := trekState.CurrentAllocation()
				ip := currentAllocation.IP()
				taskResource := currentAllocation.allocation.TaskResources[task.Name]
				fmt.Fprintf(view, "* Name: %s\n", task.Name)
				fmt.Fprintf(view, "* Node Name: %s\n", currentAllocation.node.Name)
				fmt.Fprintf(view, "* Node IP: %s\n", ip)
				fmt.Fprintf(view, "* Driver: %s\n", task.Driver)
				for k, v := range task.Config {
					fmt.Fprintf(view, "\t* %s: %+v\n", k, v)
				}
				if len(task.Env) > 0 {
					fmt.Fprintf(view, "* Env:\n")
					for k, v := range task.Env {
						fmt.Fprintf(view, "\t* %s: %s\n", k, v)
					}
				}
				for _, network := range taskResource.Networks {
					if len(network.DynamicPorts) > 0 {

						fmt.Fprintf(view, "* Dynamic Ports: ")
						for idx, dynPort := range network.DynamicPorts {
							if idx > 0 {
								fmt.Fprintf(view, ", ")
							}
							fmt.Fprintf(view, "%d (%s)", dynPort.Value, dynPort.Label)
						}
						fmt.Fprintf(view, "\n")
					}
					if len(network.ReservedPorts) > 0 {
						fmt.Fprintf(view, "* Reserved Ports: ")
						for idx, dynPort := range network.ReservedPorts {
							if idx > 0 {
								fmt.Fprintf(view, ", ")
							}
							fmt.Fprintf(view, "%d (%s)", dynPort.Value, dynPort.Label)
						}
						fmt.Fprintf(view, "\n")
					}
				}
				// if(trekState.debugModeEnabled) {
				// val := reflect.Indirect(reflect.ValueOf(task))
				// valType := val.Type()

				// for i := 0; i < val.NumField(); i++ {
				// 	field := valType.Field(i)
				// 	value := val.FieldByName(field.Name).Interface()
				// 	name := field.Name

				// 	fmt.Fprintf(view, "%s: %+v\n", name, value)
				// }
				// }

				return nil
			},
		},
		trekState,
	)
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

var bindings = []binding{
	binding{panelName: "Clusters", key: gocui.KeyEnter, handler: selectCluster},
	binding{panelName: "Clusters", key: gocui.KeyArrowRight, handler: selectCluster},
	binding{panelName: "Clusters", key: gocui.KeyArrowDown, handler: cursorDown(
		func(trekState *trekStateType, position cursorPosition) { trekState.selectedClusterIndex = position.y },
		func(trekState *trekStateType) int { return len(*trekState.nomadConnectConfiguration.Environments) })},
	binding{panelName: "Clusters", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedClusterIndex = position.y
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
		func(trekState *trekStateType) int { return len(trekState.Jobs()) })},

	binding{panelName: "Task Groups", key: gocui.KeyArrowLeft,
		handler: deleteView("Task Groups", "Jobs", func(trekState *trekStateType) { trekState.selectedAllocationGroup = 0 })},
	binding{panelName: "Task Groups", key: gocui.KeyEnter, handler: selectTaskGroup},
	binding{panelName: "Task Groups", key: gocui.KeyArrowRight, handler: selectTaskGroup},
	binding{panelName: "Task Groups", key: gocui.KeyArrowDown, handler: cursorDown(
		func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedAllocationGroup = position.y
		},
		func(trekState *trekStateType) int { return len(trekState.CurrentJob().TaskGroups) })},
	binding{panelName: "Task Groups", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedAllocationGroup = position.y
		})},

	binding{panelName: "Allocations", key: gocui.KeyArrowLeft,
		handler: deleteView("Allocations", "Task Groups", func(trekState *trekStateType) {
			trekState.selectedAllocationIndex = 0
		})},
	binding{panelName: "Allocations", key: gocui.KeyEnter, handler: selectAllocation},
	binding{panelName: "Allocations", key: gocui.KeyArrowRight, handler: selectAllocation},
	binding{panelName: "Allocations", key: gocui.KeyArrowDown,
		handler: cursorDown(
			func(trekState *trekStateType, position cursorPosition) {
				trekState.selectedAllocationIndex = position.y
			},
			func(trekState *trekStateType) int {
				return len(trekState.CurrentAllocations())
			})},
	binding{panelName: "Allocations", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedAllocationIndex = position.y
		})},

	binding{panelName: "Tasks", key: gocui.KeyArrowLeft,
		handler: deleteView("Tasks", "Allocations", func(trekState *trekStateType) { trekState.selectedTask = 0 })},
	binding{panelName: "Tasks", key: gocui.KeyEnter, handler: selectTask},
	binding{panelName: "Tasks", key: gocui.KeyArrowRight, handler: selectTask},
	binding{panelName: "Tasks", key: gocui.KeyArrowDown,
		handler: cursorDown(
			func(trekState *trekStateType, position cursorPosition) { trekState.selectedTask = position.y },
			func(trekState *trekStateType) int {
				return len(trekState.CurrentJob().TaskGroups[trekState.selectedAllocationGroup].Tasks)
			})},
	binding{panelName: "Tasks", key: gocui.KeyArrowUp,
		handler: cursorUp(func(trekState *trekStateType, position cursorPosition) {
			trekState.selectedTask = position.y
		})},

	binding{panelName: "Task", key: gocui.KeyEnter,
		handler: deleteView("Task", "Tasks", func(trekState *trekStateType) {})},

	binding{panelName: "", key: gocui.KeyCtrlC, handler: quit},
	binding{panelName: "", key: gocui.KeyF12, handler: quit},
	binding{panelName: "msg", key: gocui.KeyEnter,
		handler: deleteView("msg", "Allocations", func(trekState *trekStateType) {})},
}

func keybindings(g *gocui.Gui, trekState *trekStateType) error {
	for _, binding := range bindings {
		if err := g.SetKeybinding(binding.panelName, binding.key, gocui.ModNone, stateify(binding.handler, trekState)); err != nil {
			return err
		}
	}

	return nil
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

var initialized = false

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
		menuItems := []string{"F1:DEBUG", "F12:EXIT"}

		if v, err := g.SetView("menu_items", startX+offset, startY, endX, endY); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Highlight = false
			v.Frame = false
			v.BgColor = gocui.ColorGreen
			v.FgColor = gocui.ColorBlack

			for index, optionName := range menuItems {
				if index > 0 {
					fmt.Fprintf(v, " | ")
				}
				fmt.Fprintf(v, "%s", optionName)
			}
		}

		initialized = true
		return createView(g,
			trekView{
				name:                    "Clusters",
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
						// Can't find configuration file, applying default configuration
						address := os.Getenv("NOMAD_ADDR")
						if address == "" {
							// Defaulting on localhost
							address = "http://localhost:4646"
						}
						trekState.nomadConnectConfiguration.addEnvironment("default", address)
					} else {

						decoder := json.NewDecoder(file)
						trekState.nomadConnectConfiguration = configuration{}
						err = decoder.Decode(&trekState.nomadConnectConfiguration)
						if err != nil {
							log.Panicln(err)
						}

					}

					for _, env := range *trekState.nomadConnectConfiguration.Environments {
						fmt.Fprintf(view, "%s\n", (env).Name)
					}

					return nil
				},
			},
			trekState,
		)
	}
}

func runUI(options trekOptions) {
	trekState := new(trekStateType)

	// build ui
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = false

	g.SetManagerFunc(layout(trekState))

	if err := keybindings(g, trekState); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
