package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func mainOld() {
	app := tview.NewApplication()
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	mainContent := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Main content let's see what you put in fella")

	// NEW: Adding the inputField
	inputField := tview.NewInputField().
		SetLabel("Enter value: ").
		// SetFieldWidth(0).
		SetFieldTextColor(tcell.ColorBlack)
	// SetFieldBackgroundColor(tcell.ColorWhite).
	inputField.SetChangedFunc(func(text string) {
		mainContent.SetText(text)
	})
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		inputField.SetText("")
	})

	// NEW: Adding the flex container
	header := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(newPrimitive("Header"), 2, 0, false).
		AddItem(inputField, 0, 1, true)

	menu := newPrimitive("Menu")
	sideBar := newPrimitive("Side Bar")

	grid := tview.NewGrid().
		SetRows(3, 0, 6).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
		AddItem(mainContent, 1, 0, 1, 3, 0, 0, false).
		AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
		AddItem(mainContent, 1, 1, 1, 1, 0, 100, false).
		AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

	if err := app.SetRoot(grid, true).SetFocus(inputField).Run(); err != nil {
		panic(err)
	}
}
