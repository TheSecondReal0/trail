package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// --- Data Structures ---

type Project struct {
	Name  string
	Tasks map[string][]Entry
}

type Entry struct {
	Date    time.Time
	Content string
}

type TrailData struct {
	Projects map[string]Project
}

// --- Helpers ---

func vimList(list *tview.List) *tview.List {
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		return event
	})
	return list
}

func defaultText(text string) *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

func renderDaySummary(date time.Time, data *TrailData) string {
	var sb strings.Builder

	projectNames := make([]string, 0, len(data.Projects))
	for name := range data.Projects {
		projectNames = append(projectNames, name)
	}
	sort.Strings(projectNames)

	for _, projectName := range projectNames {
		project := data.Projects[projectName]
		taskNames := make([]string, 0, len(project.Tasks))
		for name := range project.Tasks {
			taskNames = append(taskNames, name)
		}
		sort.Strings(taskNames)

		var projectLines strings.Builder
		for _, taskName := range taskNames {
			entries := project.Tasks[taskName]
			var taskLines strings.Builder
			for _, entry := range entries {
				if entry.Date.Year() == date.Year() && entry.Date.YearDay() == date.YearDay() {
					fmt.Fprintf(&taskLines, "    %s\n", entry.Content)
				}
			}
			if taskLines.Len() > 0 {
				fmt.Fprintf(&projectLines, "  +%s\n", taskName)
				projectLines.WriteString(taskLines.String())
			}
		}
		if projectLines.Len() > 0 {
			fmt.Fprintf(&sb, "@%s\n", projectName)
			sb.WriteString(projectLines.String())
		}
	}

	result := sb.String()
	if result == "" {
		return "(no entries for this day)"
	}
	return result
}

func renderRecentSummary(days int, data *TrailData) string {
	if days <= 0 {
		return ""
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	cutoff := today.AddDate(0, 0, -(days - 1))

	projectNames := make([]string, 0, len(data.Projects))
	for name := range data.Projects {
		projectNames = append(projectNames, name)
	}
	sort.Strings(projectNames)

	var sb strings.Builder
	for _, projectName := range projectNames {
		project := data.Projects[projectName]
		taskNames := make([]string, 0, len(project.Tasks))
		for name := range project.Tasks {
			taskNames = append(taskNames, name)
		}
		sort.Strings(taskNames)

		var projectLines strings.Builder
		for _, taskName := range taskNames {
			dateMap := make(map[time.Time][]string)
			for _, entry := range project.Tasks[taskName] {
				if !entry.Date.Before(cutoff) && !entry.Date.After(today) {
					dateMap[entry.Date] = append(dateMap[entry.Date], entry.Content)
				}
			}
			if len(dateMap) == 0 {
				continue
			}

			dates := make([]time.Time, 0, len(dateMap))
			for d := range dateMap {
				dates = append(dates, d)
			}
			sort.Slice(dates, func(i, j int) bool {
				return dates[i].After(dates[j])
			})

			fmt.Fprintf(&projectLines, "  +%s\n", taskName)
			for _, date := range dates {
				fmt.Fprintf(&projectLines, "    %s\n", date.Format("2006-01-02"))
				for _, content := range dateMap[date] {
					fmt.Fprintf(&projectLines, "      %s\n", content)
				}
			}
		}

		if projectLines.Len() > 0 {
			fmt.Fprintf(&sb, "@%s\n", projectName)
			sb.WriteString(projectLines.String())
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

var screenNames = []string{"projects", "tasks", "days", "recent"}

func switchScreen(pages *tview.Pages, current *string, direction int) {
	idx := 0
	for i, name := range screenNames {
		if name == *current {
			idx = i
			break
		}
	}
	idx = (idx + direction + len(screenNames)) % len(screenNames)
	*current = screenNames[idx]
	pages.SwitchToPage(*current)
}

// --- ProjectsScreen ---

type ProjectsScreen struct {
	Root           *tview.Grid
	innerPages     *tview.Pages
	filter         *tview.InputField
	list           *tview.List
	taskList       *tview.List
	taskContent    *tview.TextView
	data           *TrailData
	app            *tview.Application
	currentProject *Project
}

func newProjectsScreen(data *TrailData, app *tview.Application) *ProjectsScreen {
	ps := &ProjectsScreen{data: data, app: app}

	ps.filter = tview.NewInputField().
		SetLabel("Filter Projects: ").
		SetFieldTextColor(tcell.ColorBlack).
		SetChangedFunc(func(text string) {
			ps.populateProjects(text)
		})
	ps.filter.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.SetFocus(ps.list)
		}
	})

	ps.list = vimList(tview.NewList())
	ps.taskList = vimList(tview.NewList())
	ps.taskContent = tview.NewTextView().SetScrollable(true)

	ps.innerPages = tview.NewPages()
	ps.innerPages.AddPage("list", ps.list, true, true)
	ps.innerPages.AddPage("tasks", ps.taskList, true, false)
	ps.innerPages.AddPage("content", ps.taskContent, true, false)

	ps.Root = tview.NewGrid().SetRows(1, 0).SetColumns(0).SetBorders(true)
	ps.Root.AddItem(ps.filter, 0, 0, 1, 1, 0, 0, false)
	ps.Root.AddItem(ps.innerPages, 1, 0, 1, 1, 0, 0, true)

	ps.populateProjects("")
	return ps
}

func (ps *ProjectsScreen) populateProjects(filter string) {
	ps.list.Clear()
	names := make([]string, 0, len(ps.data.Projects))
	for name := range ps.data.Projects {
		if filter == "" || strings.Contains(name, filter) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		p := ps.data.Projects[name]
		ps.list.AddItem(p.Name, "", 0, func() {
			ps.showTasks(p)
		})
	}
}

func (ps *ProjectsScreen) showTasks(project Project) {
	ps.currentProject = &project
	ps.taskList.Clear()

	taskNames := make([]string, 0, len(project.Tasks))
	for name := range project.Tasks {
		taskNames = append(taskNames, name)
	}
	sort.Strings(taskNames)

	for _, name := range taskNames {
		entries := slices.Clone(project.Tasks[name])
		slices.SortStableFunc(entries, func(a, b Entry) int {
			return b.Date.Compare(a.Date)
		})
		ps.taskList.AddItem(name, "", 0, func() {
			ps.showTaskContent(entries)
		})
	}
	ps.innerPages.SwitchToPage("tasks")
	ps.app.SetFocus(ps.taskList)
}

func (ps *ProjectsScreen) showTaskContent(entries []Entry) {
	if len(entries) == 0 {
		ps.taskContent.SetText("")
	} else {
		currentDate := entries[0].Date
		text := currentDate.Format("06-01-02")
		for _, entry := range entries {
			if entry.Date != currentDate {
				currentDate = entry.Date
				text += "\n" + currentDate.Format("06-01-02")
			}
			text += "\n" + entry.Content
		}
		ps.taskContent.SetText(text)
	}
	ps.innerPages.SwitchToPage("content")
	ps.app.SetFocus(ps.taskContent)
}

func (ps *ProjectsScreen) handleEsc() {
	if ps.app.GetFocus() == ps.filter {
		ps.app.SetFocus(ps.list)
		return
	}
	name, _ := ps.innerPages.GetFrontPage()
	switch name {
	case "content":
		ps.innerPages.SwitchToPage("tasks")
		ps.app.SetFocus(ps.taskList)
	case "tasks":
		ps.currentProject = nil
		ps.innerPages.SwitchToPage("list")
		ps.app.SetFocus(ps.list)
	}
}

func (ps *ProjectsScreen) focusFilter() {
	ps.app.SetFocus(ps.filter)
}

// --- TasksScreen ---

type TasksScreen struct {
	Root       *tview.Grid
	innerPages *tview.Pages
	filter     *tview.InputField
	list       *tview.List
	content    *tview.TextView
	data       *TrailData
	app        *tview.Application
}

func newTasksScreen(data *TrailData, app *tview.Application) *TasksScreen {
	ts := &TasksScreen{data: data, app: app}

	ts.filter = tview.NewInputField().
		SetLabel("Filter Tasks: ").
		SetFieldTextColor(tcell.ColorBlack).
		SetChangedFunc(func(text string) {
			ts.populateTasks(text)
		})
	ts.filter.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.SetFocus(ts.list)
		}
	})

	ts.list = vimList(tview.NewList())
	ts.content = tview.NewTextView().SetScrollable(true)

	ts.innerPages = tview.NewPages()
	ts.innerPages.AddPage("list", ts.list, true, true)
	ts.innerPages.AddPage("content", ts.content, true, false)

	ts.Root = tview.NewGrid().SetRows(1, 0).SetColumns(0).SetBorders(true)
	ts.Root.AddItem(ts.filter, 0, 0, 1, 1, 0, 0, false)
	ts.Root.AddItem(ts.innerPages, 1, 0, 1, 1, 0, 0, true)

	ts.populateTasks("")
	return ts
}

func (ts *TasksScreen) populateTasks(filter string) {
	ts.list.Clear()

	type taskItem struct {
		label   string
		entries []Entry
	}

	var items []taskItem
	for _, project := range ts.data.Projects {
		for taskName, entries := range project.Tasks {
			label := project.Name + "/" + taskName
			if filter == "" || strings.Contains(label, filter) {
				items = append(items, taskItem{label: label, entries: entries})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].label < items[j].label
	})

	for _, item := range items {
		entries := slices.Clone(item.entries)
		slices.SortStableFunc(entries, func(a, b Entry) int {
			return b.Date.Compare(a.Date)
		})
		label := item.label
		ts.list.AddItem(label, "", 0, func() {
			ts.showContent(entries)
		})
	}
}

func (ts *TasksScreen) showContent(entries []Entry) {
	if len(entries) == 0 {
		ts.content.SetText("")
	} else {
		currentDate := entries[0].Date
		text := currentDate.Format("06-01-02")
		for _, entry := range entries {
			if entry.Date != currentDate {
				currentDate = entry.Date
				text += "\n" + currentDate.Format("06-01-02")
			}
			text += "\n" + entry.Content
		}
		ts.content.SetText(text)
	}
	ts.innerPages.SwitchToPage("content")
	ts.app.SetFocus(ts.content)
}

func (ts *TasksScreen) handleEsc() {
	if ts.app.GetFocus() == ts.filter {
		ts.app.SetFocus(ts.list)
		return
	}
	name, _ := ts.innerPages.GetFrontPage()
	if name == "content" {
		ts.innerPages.SwitchToPage("list")
		ts.app.SetFocus(ts.list)
	}
}

func (ts *TasksScreen) focusFilter() {
	ts.app.SetFocus(ts.filter)
}

// --- DaysScreen ---

type DaysScreen struct {
	Root       *tview.Grid
	innerPages *tview.Pages
	filter     *tview.InputField
	list       *tview.List
	detail     *tview.TextView
	data       *TrailData
	app        *tview.Application
}

func newDaysScreen(data *TrailData, app *tview.Application) *DaysScreen {
	ds := &DaysScreen{data: data, app: app}

	ds.filter = tview.NewInputField().
		SetLabel("Filter Dates: ").
		SetFieldTextColor(tcell.ColorBlack).
		SetChangedFunc(func(text string) {
			ds.populateDays(text)
		})
	ds.filter.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.SetFocus(ds.list)
		}
	})

	ds.list = vimList(tview.NewList())
	ds.detail = tview.NewTextView().SetScrollable(true)

	ds.innerPages = tview.NewPages()
	ds.innerPages.AddPage("list", ds.list, true, true)
	ds.innerPages.AddPage("detail", ds.detail, true, false)

	ds.Root = tview.NewGrid().SetRows(1, 0).SetColumns(0).SetBorders(true)
	ds.Root.AddItem(ds.filter, 0, 0, 1, 1, 0, 0, false)
	ds.Root.AddItem(ds.innerPages, 1, 0, 1, 1, 0, 0, true)

	ds.populateDays("")
	return ds
}

func (ds *DaysScreen) populateDays(filter string) {
	ds.list.Clear()

	dateSet := make(map[time.Time]struct{})
	for _, project := range ds.data.Projects {
		for _, entries := range project.Tasks {
			for _, entry := range entries {
				dateSet[entry.Date] = struct{}{}
			}
		}
	}

	dates := make([]time.Time, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].After(dates[j])
	})

	for _, date := range dates {
		label := date.Format("2006-01-02")
		if filter != "" && !strings.Contains(label, filter) {
			continue
		}
		d := date
		ds.list.AddItem(label, "", 0, func() {
			ds.showDetail(d)
		})
	}
}

func (ds *DaysScreen) showDetail(date time.Time) {
	summary := renderDaySummary(date, ds.data)
	ds.detail.SetText(summary)
	ds.innerPages.SwitchToPage("detail")
	ds.app.SetFocus(ds.detail)
}

func (ds *DaysScreen) handleEsc() {
	if ds.app.GetFocus() == ds.filter {
		ds.app.SetFocus(ds.list)
		return
	}
	name, _ := ds.innerPages.GetFrontPage()
	if name == "detail" {
		ds.innerPages.SwitchToPage("list")
		ds.app.SetFocus(ds.list)
	}
}

func (ds *DaysScreen) focusFilter() {
	ds.app.SetFocus(ds.filter)
}

// --- RecentScreen ---

type RecentScreen struct {
	Root    *tview.Grid
	days    *tview.InputField
	content *tview.TextView
	data    *TrailData
	app     *tview.Application
}

func newRecentScreen(data *TrailData, app *tview.Application) *RecentScreen {
	rs := &RecentScreen{data: data, app: app}

	rs.content = tview.NewTextView().SetScrollable(true)

	rs.days = tview.NewInputField().
		SetLabel("Last N days: ").
		SetText("28").
		SetFieldTextColor(tcell.ColorBlack).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetChangedFunc(func(text string) {
			n, err := strconv.Atoi(text)
			if err != nil || n <= 0 {
				rs.content.SetText("")
				return
			}
			rs.content.SetText(renderRecentSummary(n, data))
		})
	rs.days.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.SetFocus(rs.content)
		}
	})

	rs.Root = tview.NewGrid().SetRows(1, 0).SetColumns(0).SetBorders(true)
	rs.Root.AddItem(rs.days, 0, 0, 1, 1, 0, 0, false)
	rs.Root.AddItem(rs.content, 1, 0, 1, 1, 0, 0, true)

	rs.content.SetText(renderRecentSummary(28, data))
	return rs
}

func (rs *RecentScreen) handleEsc() {
	if rs.app.GetFocus() == rs.days {
		rs.app.SetFocus(rs.content)
	}
}

func (rs *RecentScreen) focusFilter() {
	rs.app.SetFocus(rs.days)
}

// --- Main ---

func main() {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		stateDir = filepath.Join(home, ".local", "state")
	}
	logDir := filepath.Join(stateDir, "trail")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}
	logPath := filepath.Join(logDir, "trail.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	app := tview.NewApplication()

	notesDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projects := ProjectsFromDirectory(notesDir)
	trailData := TrailData{Projects: projects}

	rootPages := tview.NewPages()
	currentScreen := "projects"

	ps := newProjectsScreen(&trailData, app)
	ts := newTasksScreen(&trailData, app)
	ds := newDaysScreen(&trailData, app)
	rs := newRecentScreen(&trailData, app)

	rootPages.AddPage("projects", ps.Root, true, true)
	rootPages.AddPage("tasks", ts.Root, true, false)
	rootPages.AddPage("days", ds.Root, true, false)
	rootPages.AddPage("recent", rs.Root, true, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			switchScreen(rootPages, &currentScreen, 1)
			return nil
		case tcell.KeyBacktab:
			switchScreen(rootPages, &currentScreen, -1)
			return nil
		case tcell.KeyEscape:
			switch currentScreen {
			case "projects":
				ps.handleEsc()
			case "tasks":
				ts.handleEsc()
			case "days":
				ds.handleEsc()
			case "recent":
				rs.handleEsc()
			}
			return nil
		}
		if event.Rune() == '/' {
			if _, ok := app.GetFocus().(*tview.InputField); !ok {
				switch currentScreen {
				case "projects":
					ps.focusFilter()
				case "tasks":
					ts.focusFilter()
				case "days":
					ds.focusFilter()
				case "recent":
					rs.focusFilter()
				}
				return nil
			}
		}
		return event
	})

	if err := app.SetRoot(rootPages, true).SetFocus(rootPages).Run(); err != nil {
		panic(err)
	}
}
