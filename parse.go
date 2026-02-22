package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	// "path/filepath"
)

func parseDirectory(dir string) []Project {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// TODO log error to file
		panic(err)
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			// TODO log error to file
			panic(err)
		}
		// TODO should I save whole path, is this name only name of file?
		files = append(files, info.Name())
	}

	return nil
}
func ProjectsFromFile(path string) map[string]Project {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer file.Close()

	projectMap := make(map[string]Project)

	var currentProject *string
	var currentTask *string

	projectRegex, _ := regexp.Compile(`(?:^|\s)@([a-zA-Z0-9_-]+)`)
	taskRegex, _ := regexp.Compile(`(?:^|\s)\+([a-zA-Z0-9_-]+)`)
	entryRegex, _ := regexp.Compile(`^(?:\*|-|\s)`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		log.Println("Scanning line: ", text)
		// if we are in an entry or new heading
		if entryRegex.MatchString(text) {
			log.Println("line is entry")
			// expect currentProject and currentTask are already set... ignore otherwise
			if (currentProject == nil) || (currentTask == nil) {
				continue
			}
			log.Println("Adding entry under project: ", *currentProject, ", task: ", *currentTask)

			// TODO maybe some sanitization here to take out bullets/dashes etc.
			// TODO handle TODOS
			entry := Entry{
				Date:    path,
				Content: text,
			}
			projectMap[*currentProject].Tasks[*currentTask] = append(projectMap[*currentProject].Tasks[*currentTask], entry)
		} else {
			log.Println("line is new heading")
			projectMatches := projectRegex.FindAllStringSubmatch(text, -1)
			if len(projectMatches) == 0 {
				log.Println("No project matches")
				continue
			}
			projectMatch := projectMatches[0][1]
			taskMatches := taskRegex.FindAllStringSubmatch(text, -1)
			if len(taskMatches) == 0 {
				log.Println("No task matches")
				continue
			}
			taskMatch := taskMatches[0][1]

			// ignoring sections without project AND task for now

			currentProject = &projectMatch
			currentTask = &taskMatch
			_, ok := projectMap[*currentProject]
			if !ok {
				projectMap[*currentProject] = Project{
					Name:  *currentProject,
					Tasks: make(map[string][]Entry),
				}
			}
			_, ok = projectMap[*currentProject].Tasks[*currentTask]
			if !ok {
				projectMap[*currentProject].Tasks[*currentTask] = make([]Entry, 0)
			}
		}
	}

	return projectMap
}
