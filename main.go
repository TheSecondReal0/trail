package main

import (
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Project struct {
	Name  string
	Tasks map[string][]Entry
}

type Entry struct {
	Date    time.Time
	Content string
}

type TrailState int

const (
	ViewProjects TrailState = iota
	ViewTasksOfProject
	ViewDays
	ViewContents
)

type TrailData struct {
	Projects map[string]Project
}

type ProjectView struct {
	Root            *tview.Grid
	filter          *tview.InputField
	list            *tview.List
	trailData       *TrailData
	projectSelected func(Project)
}

func newProjectView(trailData *TrailData, projectSelected func(Project)) (pv ProjectView) {
	pv.Root = tview.NewGrid()
	pv.filter = tview.NewInputField().SetLabel("Filter Projects: ").SetFieldTextColor(tcell.ColorBlack).SetChangedFunc(func(text string) {
		pv.filterProjects(text)
	})
	pv.list = tview.NewList()
	pv.Root.AddItem(pv.filter, 0, 0, 1, 1, 0, 0, false)
	pv.Root.AddItem(pv.list, 1, 0, 1, 1, 0, 0, true)
	pv.trailData = trailData
	pv.projectSelected = projectSelected
	pv.showAllProjects()
	return pv
}

func (pv ProjectView) showAllProjects() {
	pv.list.Clear()
	for _, p := range (*pv.trailData).Projects {
		pv.list.AddItem(p.Name, "", 0, func() {
			pv.projectSelected(p)
		})
	}
}

func (pv ProjectView) filterProjects(filter string) {
	pv.list.Clear()
	for _, p := range (*pv.trailData).Projects {
		if !strings.Contains(p.Name, filter) {
			continue
		}
		pv.list.AddItem(p.Name, "", 0, func() {
			pv.projectSelected(p)
		})
	}
}

func defaultText(text string) *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

func main() {
	logFile, err := os.OpenFile("trail.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := tview.NewApplication()
	state := ViewProjects
	var currentProject *Project = nil
	// var currentTask *string = nil

	projects := ProjectsFromDirectory("/home/asa/dev/trail/notes/")
	trailData := TrailData{
		Projects: projects,
	}

	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		SetBorders(true)

	list := tview.NewList()
	taskContents := tview.NewTextView()

	showTaskContents := func(contents string) {
		state = ViewContents
		taskContents.SetText(contents)
		grid.RemoveItem(list)
		grid.AddItem(taskContents, 0, 0, 1, 1, 0, 0, false)
	}
	showTasks := func(project Project) {
		state = ViewTasksOfProject
		// currentTask = nil
		list.Clear()
		for name, entries := range project.Tasks {
			slices.SortStableFunc(entries, func(a, b Entry) int {
				return b.Date.Compare(a.Date)
			})

			list.AddItem(name, "", 0, func() {
				if len(entries) == 0 {
					showTaskContents("")
					return
				}
				currentDate := entries[0].Date
				text := currentDate.Format("06-01-02")
				for _, entry := range entries {
					if entry.Date != currentDate {
						currentDate = entry.Date
						text += "\n" + currentDate.Format("06-01-02")
					}
					text += "\n" + entry.Content

				}
				showTaskContents(text)
			})
		}
		grid.RemoveItem(taskContents)
		grid.AddItem(list, 0, 0, 1, 1, 0, 0, false)
	}

	projectView := newProjectView(&trailData, func(p Project) {
		currentProject = &p
		showTasks(p)
	})
	showProjects := func() {
		state = ViewProjects
		currentProject = nil
		projectView.showAllProjects()
		grid.RemoveItem(taskContents)
		grid.AddItem(projectView.Root, 0, 0, 1, 1, 0, 0, false)
		app.SetFocus(projectView.list)
	}

	showProjects()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			switch state {
			case ViewTasksOfProject:
				showProjects()
			case ViewContents:
				showTasks(*currentProject)
			}
			return nil
		}
		if event.Rune() == '/' {
			switch state {
			case ViewProjects:
				app.SetFocus(projectView.filter)
				return nil
				// case ViewTasksOfProject:
				// 	showProjects()
				// case ViewContents:
				// 	showTasks(*currentProject)
			}
		}
		return event
	})

	// search := tview.NewInputField().
	// 	SetLabel("Search").
	// 	SetFieldTextColor(tcell.ColorBlack)
	// search.SetChangedFunc(func(text string) {
	// })
	// search.SetDoneFunc(func(key tcell.Key) {
	// 	if key != tcell.KeyEnter {
	// 		return
	// 	}
	// 	search.SetText("")
	// })

	if err := app.SetRoot(grid, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}
