package main

import (
	"log"
	"os"
	"slices"
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
	ViewTasks
	ViewContents
)

func defaultText(text string) *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

func main() {
	logFile, err := os.OpenFile("trail.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
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

	grid := tview.NewGrid().
		SetRows(1, 0).
		SetColumns(0) //.
		// SetBorders(true)

	grid.AddItem(tview.NewTextView().SetText("Trail"), 0, 0, 1, 1, 0, 0, false)

	mainContent := tview.NewFlex()
	mainContent.SetBorder(true)
	grid.AddItem(mainContent, 1, 0, 1, 1, 0, 0, false)

	list := tview.NewList()
	list.SetHighlightFullLine(true)
	taskContents := tview.NewTextView()

	showTaskContents := func(contents string) {
		state = ViewContents
		taskContents.SetText(contents)
		mainContent.Clear()
		mainContent.AddItem(taskContents, 0, 1, false)
	}
	showTasks := func(project Project) {
		state = ViewTasks
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
		mainContent.Clear()
		mainContent.AddItem(list, 0, 1, false)
	}
	showProjects := func() {
		state = ViewProjects
		currentProject = nil
		list.Clear()
		for _, p := range projects {
			list.AddItem(p.Name, "", 0, func() {
				currentProject = &p
				showTasks(p)
			})
		}
		mainContent.Clear()
		mainContent.AddItem(list, 0, 1, false)
	}

	showProjects()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() != tcell.KeyEscape {
			return event
		}
		switch state {
		case ViewTasks:
			showProjects()
		case ViewContents:
			showTasks(*currentProject)
		}
		return nil
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
