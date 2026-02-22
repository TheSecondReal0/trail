package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Project struct {
	Name  string
	Tasks map[string]string
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
	app := tview.NewApplication()
	state := ViewProjects
	var currentProject *Project = nil
	// var currentTask *string = nil

	projects := []Project{
		Project{
			Name: "trail",
			Tasks: map[string]string{
				"ui":         "* Added some good new things to this bad boi\n* Tried to get layout better\n* Not sure how I want the search to work...",
				"build":      "* Not sure how to build this bad boi",
				"versioning": "* [ ] Should get cog in here and some automations to do it for me",
			},
		},
		Project{
			Name: "asaio-strategy",
			Tasks: map[string]string{
				"c#-conversion": "* Converted rest of stuff to native C#\n* [ ] Need to fix notifications\n* TODO: Need to fix storehouses make them selectable",
				"ui":            "* TODO: Figure out UI for unit automation",
			},
		},
	}

	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		SetBorders(true)

	list := tview.NewList()
	taskContents := defaultText("")

	showTaskContents := func(contents string) {
		state = ViewContents
		taskContents.SetText(contents)
		grid.RemoveItem(list)
		grid.AddItem(taskContents, 0, 0, 1, 1, 0, 0, false)
	}
	showTasks := func(project Project) {
		state = ViewTasks
		// currentTask = nil
		list.Clear()
		for name, contents := range project.Tasks {
			list.AddItem(name, "", 0, func() {
				// currentTask = &name
				showTaskContents(contents)
			})
		}
		grid.RemoveItem(taskContents)
		grid.AddItem(list, 0, 0, 1, 1, 0, 0, false)
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
		grid.RemoveItem(taskContents)
		grid.AddItem(list, 0, 0, 1, 1, 0, 0, false)
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
