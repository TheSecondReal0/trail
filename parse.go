package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
	// "path/filepath"
)

func ProjectsFromDirectory(dir string) map[string]Project {
	projectMap := make(map[string]Project)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
		// TODO should I save whole path, is this name only name of file?
		files = append(files, info.Name())
		ProjectsFromFile(filepath.Join(dir, info.Name()), projectMap)
	}

	return projectMap
}

// Optionally pass in existing projectMap to add onto it, nil if want new map
func ProjectsFromFile(path string, projectMap map[string]Project) map[string]Project {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer file.Close()

	if projectMap == nil {
		projectMap = make(map[string]Project)
	}

	var currentProject *string
	var currentTask *string

	projectRegex, _ := regexp.Compile(`(?:^|\s)@([a-zA-Z0-9_.-]+)`)
	taskRegex, _ := regexp.Compile(`(?:^|\s)\+([a-zA-Z0-9_.-]+)`)
	entryRegex, _ := regexp.Compile(`^(?:\*|-|\s)`)
	dateRegex, _ := regexp.Compile(`(\d\d-\d\d-\d\d)`)
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
			dateMatches := dateRegex.FindAllStringSubmatch(filepath.Base(path), -1)
			if len(dateMatches) == 0 {
				log.Println("No date matches")
				continue
			}
			dateMatch := dateMatches[0][1]
			dateTime, _ := time.Parse("06-01-02", dateMatch)

			// TODO maybe some sanitization here to take out bullets/dashes etc.
			// TODO handle TODOS
			entry := Entry{
				Date:    dateTime,
				Content: text,
			}
			projectMap[*currentProject].Tasks[*currentTask] = append(projectMap[*currentProject].Tasks[*currentTask], entry)
		} else {
			log.Println("line is new heading")
			// should not append entries found after another header to the last project/task
			currentProject = nil
			currentTask = nil
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
