package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/nomad/api"
	"github.com/jroimartin/gocui"
)

type configuration struct {
	Environments []environment
}
type environment struct {
	Name    string
	Address string
}

var selectedCluster int = 0
var selectedJob int = 0
var selectedTaskGroup int = 0
var selectedTask int = 0
var selectedService int = 0
var config *api.Config
var client *api.Client
var jobs []api.Job
var jobsHandle *api.Jobs
var options *api.QueryOptions
var nomadConnectConfiguration configuration

func clustersViewCursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()

		// Prevent scrolling past clusters
		if v.Title == "Clusters" {
			numClusters := len(nomadConnectConfiguration.Environments)
			if cy < 0 || cy >= numClusters-1 {
				return nil
			}
		}

		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func clustersViewCursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func jobsViewCursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
		selectedJob = cy - 1
	}
	return nil
}
func jobsViewCursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()

		// Prevent scrolling past jobs
		if v.Title == "Jobs" {
			numJobs := len(jobs)
			if cy < 0 || cy >= numJobs-1 {
				return nil
			}
		}

		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
		selectedJob = cy + 1
	}
	return nil
}

func taskGroupsViewCursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
		selectedTaskGroup = cy - 1
	}
	return nil
}
func taskGroupsViewCursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()

		// Prevent scrolling past jobs
		if v.Title == "Task Groups" {
			numGroups := len(jobs[selectedJob].TaskGroups)
			if cy < 0 || cy >= numGroups-1 {
				return nil
			}
		}

		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
		selectedTaskGroup = cy + 1
	}
	return nil
}

func tasksViewCursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
		selectedTask = cy - 1
	}
	return nil
}
func tasksViewCursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()

		// Prevent scrolling past tasks
		if v.Title == "Tasks" {
			numTasks := len(jobs[selectedJob].TaskGroups[selectedTaskGroup].Tasks)
			if cy < 0 || cy >= numTasks-1 {
				return nil
			}
		}

		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
		selectedTask = cy + 1
	}
	return nil
}

func confirmTaskSelection(g *gocui.Gui, v *gocui.View) error {
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

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func selectCluster(g *gocui.Gui, v *gocui.View) error {
	_, maxY := g.Size()
	if v, err := g.SetView("Jobs", 30, 2, 60, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Jobs"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		jobsHandle = client.Jobs()
		jobListStubs, _, _ := jobsHandle.List(options)
		jobs = make([]api.Job, 0)
		for _, job := range jobListStubs {
			fullJob, _, _ := jobsHandle.Info(job.ID, options)
			jobs = append(jobs, *fullJob)
			fmt.Fprintf(v, "%s (%s)\n", *(fullJob.Name), *(fullJob.ID))
		}
		v.Editable = false
		v.Wrap = true
	}
	if _, err := g.SetCurrentView("Jobs"); err != nil {
		return err
	}
	return nil
}

func selectJob(g *gocui.Gui, v *gocui.View) error {
	_, maxY := g.Size()

	if len(jobs) < 1 {
		return nil
	}

	job := jobs[selectedJob]

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
		v.Wrap = true
	}
	if _, err := g.SetCurrentView("Task Groups"); err != nil {
		return err
	}
	return nil
}
func selectTaskGroup(g *gocui.Gui, v *gocui.View) error {
	_, maxY := g.Size()
	if v, err := g.SetView("Tasks", 90, 2, 120, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Tasks"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		taskGroup := jobs[selectedJob].TaskGroups[selectedTaskGroup]

		for _, task := range taskGroup.Tasks {
			fmt.Fprintf(v, "%s (%s)\n", (task.Name), (task.Driver))
		}

		v.Editable = false
		v.Wrap = true
	}
	if _, err := g.SetCurrentView("Tasks"); err != nil {
		return err
	}
	return nil
}

func selectTask(g *gocui.Gui, v *gocui.View) error {
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

		task := jobs[selectedJob].TaskGroups[selectedTaskGroup].Tasks[selectedTask]

		for _, service := range task.Services {
			fmt.Fprintf(v, "%s\n", (service.Name))
		}

		v.Editable = false
		v.Wrap = true
	}
	if _, err := g.SetCurrentView("Services"); err != nil {
		return err
	}
	return nil
}

func clearJobsView(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("Jobs"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("Clusters"); err != nil {
		return err
	}
	selectedJob = 0
	return nil
}

func clearTaskGroupsView(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("Task Groups"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("Jobs"); err != nil {
		return err
	}
	selectedTaskGroup = 0
	return nil
}

func clearTasksView(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("Tasks"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("Task Groups"); err != nil {
		return err
	}
	selectedTask = 0
	return nil
}

func clearServicesView(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("Services"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("Tasks"); err != nil {
		return err
	}
	selectedService = 0
	return nil
}

// binding is some binding
type binding struct {
	panelName string
	key       gocui.Key
	handler   func(*gocui.Gui, *gocui.View) error
}

func clearConfirmation(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("Tasks"); err != nil {
		return err
	}
	return nil
}

var bindings = []binding{
	binding{panelName: "Clusters", key: gocui.KeyEnter, handler: selectCluster},
	binding{panelName: "Clusters", key: gocui.KeyArrowRight, handler: selectCluster},
	binding{panelName: "Clusters", key: gocui.KeyArrowDown, handler: clustersViewCursorDown},
	binding{panelName: "Clusters", key: gocui.KeyArrowUp, handler: clustersViewCursorUp},

	binding{panelName: "Jobs", key: gocui.KeyArrowLeft, handler: clearJobsView},
	binding{panelName: "Jobs", key: gocui.KeyEnter, handler: selectJob},
	binding{panelName: "Jobs", key: gocui.KeyArrowRight, handler: selectJob},
	binding{panelName: "Jobs", key: gocui.KeyArrowDown, handler: jobsViewCursorDown},
	binding{panelName: "Jobs", key: gocui.KeyArrowUp, handler: jobsViewCursorUp},

	binding{panelName: "Task Groups", key: gocui.KeyArrowLeft, handler: clearTaskGroupsView},
	binding{panelName: "Task Groups", key: gocui.KeyEnter, handler: selectTaskGroup},
	binding{panelName: "Task Groups", key: gocui.KeyArrowRight, handler: selectTaskGroup},
	binding{panelName: "Task Groups", key: gocui.KeyArrowDown, handler: taskGroupsViewCursorDown},
	binding{panelName: "Task Groups", key: gocui.KeyArrowUp, handler: taskGroupsViewCursorUp},

	binding{panelName: "Tasks", key: gocui.KeyArrowLeft, handler: clearTasksView},
	binding{panelName: "Tasks", key: gocui.KeyEnter, handler: selectTask},
	binding{panelName: "Tasks", key: gocui.KeyArrowRight, handler: selectTask},
	binding{panelName: "Tasks", key: gocui.KeyArrowDown, handler: tasksViewCursorDown},
	binding{panelName: "Tasks", key: gocui.KeyArrowUp, handler: tasksViewCursorUp},

	binding{panelName: "Services", key: gocui.KeyArrowLeft, handler: clearServicesView},

	binding{panelName: "", key: gocui.KeyCtrlC, handler: quit},
	binding{panelName: "msg", key: gocui.KeyEnter, handler: clearConfirmation},
}

func keybindings(g *gocui.Gui) error {
	for _, binding := range bindings {
		if err := g.SetKeybinding(binding.panelName, binding.key, gocui.ModNone, binding.handler); err != nil {
			return err
		}
	}

	return nil
}

func layout(g *gocui.Gui) error {
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
		nomadConnectConfiguration = configuration{}
		err = decoder.Decode(&nomadConnectConfiguration)
		if err != nil {
			log.Panicln(err)
		}

		for _, env := range nomadConnectConfiguration.Environments {
			fmt.Fprintf(v, "%s\n", env.Name)
		}

		if _, err := g.SetCurrentView("Clusters"); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	//connect to nomad
	config = api.DefaultConfig()
	var err error
	client, err = api.NewClient(config)
	options = &api.QueryOptions{}

	if err != nil {
		log.Panicln(err)
	}

	if len(os.Args) < 2 {
		// build ui
		g, err := gocui.NewGui(gocui.OutputNormal)
		if err != nil {
			log.Panicln(err)
		}
		defer g.Close()

		g.Cursor = true

		g.SetManagerFunc(layout)

		if err := keybindings(g); err != nil {
			log.Panicln(err)
		}

		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Panicln(err)
		}
	} else {
		jobID := os.Args[1]
		allocs := client.Allocations()
		allocsListStub, _, _ := allocs.List(options)
		found := false
		var foundAllocation *api.Allocation
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
			nodes := client.Nodes()
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
