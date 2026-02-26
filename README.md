# trail

A terminal UI for browsing work notes. Run it from a directory containing `.md` files.

## Note format

Files must be named with a date: `YY-MM-DD.md` (e.g. `25-02-14.md`). The date in the filename is used as the entry date.

A heading line must contain both a project tag (`@name`) and a task tag (`+name`). Lines below the heading that start with `*`, `-`, or whitespace are recorded as entries for that project/task pair.

```
@myproject +some-task
- worked on the thing
- fixed the other thing

@myproject +another-task
* made progress

worked on @anotherproject to finish +third-task
* finished it, yay
```

Project and task names may contain letters, digits, `_`, `.`, and `-`. A file can contain entries for multiple projects and tasks, and entries accumulate across all files in the directory.

## Screens

Navigate between screens with `Tab` / `Shift-Tab`. Press `/` to focus the filter input on any screen.

### projects

Lists all projects. Select one to drill into its tasks, then select a task to see all entries grouped by date, newest first. Press `Esc` to go back one level.

### tasks

Lists all tasks across every project as `project/task`. Select one to see its entries. Press `Esc` to return to the list.

### days

Lists every date that has at least one entry, newest first. Select a date to see all entries recorded that day, grouped by project and task. Press `Esc` to return to the list.

### recent

Shows all activity within a rolling window. The "Last N days" input controls how far back to look (default 28). Projects and tasks are drawn as nested boxes. Press `Enter` to move focus to the content area and scroll with `j`/`k` or arrow keys.

## Controls

| Key | Action |
|-----|--------|
| `Tab` / `Shift-Tab` | Next / previous screen |
| `/` | Focus filter or days input |
| `Enter` | Select item / confirm input |
| `Esc` | Go back / deselect |
| `j` / `k` | Move down / up in lists |
| Arrow keys | Move in lists and scroll content |
| `Ctrl-C` | Quit |
