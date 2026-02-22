package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	// Sample data to filter
	items := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry", "Fig", "Grape", "Honeydew", "Kiwi", "Lemon", "Mango", "Orange"}

	// Create the filterable list
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	// Function to populate/filter the list
	updateList := func(filter string) {
		list.Clear()
		for _, item := range items {
			if strings.Contains(strings.ToLower(item), strings.ToLower(filter)) {
				list.AddItem(item, "", 0, nil)
			}
		}
	}

	// Initially populate with all items
	updateList("")

	// Create dynamic footer for guidance text
	footer := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("↑/↓ Navigate list • : Filter")

	// Function to update guidance text based on focus
	updateGuidance := func(focusedOnList bool) {
		if focusedOnList {
			footer.SetText("↑/↓ Navigate list • : Filter")
		} else {
			footer.SetText("Type to filter • Enter to focus list")
		}
	}

	inputField := tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(0).
		SetFieldTextColor(tcell.ColorBlack).
		SetFieldBackgroundColor(tcell.ColorWhite)

	// Add real-time filtering as user types
	inputField.SetChangedFunc(func(text string) {
		updateList(text)
	})

	// On Enter, focus the list for navigation
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.SetFocus(list)
			updateGuidance(true)
		}
	})

	// Add list navigation - return to input on selection
	list.SetDoneFunc(func() {
		app.SetFocus(inputField)
		updateGuidance(false)
	})

	// Handle keys for list: escape to return to input, ':' to focus filter
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetFocus(inputField)
			updateGuidance(false)
			return nil
		} else if event.Rune() == ':' {
			app.SetFocus(inputField)
			updateGuidance(false)
			return nil
		}
		return event
	})

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
		AddItem(footer, 2, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
		AddItem(list, 1, 0, 1, 3, 0, 0, false).
		AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
		AddItem(list, 1, 1, 1, 1, 0, 100, false).
		AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

	if err := app.SetRoot(grid, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}
