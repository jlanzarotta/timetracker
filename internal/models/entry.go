package models

import (
	"strings"
	"timetracker/constants"
)

type Entry struct {
	Uid           int64
	Project       string
	Note          string
	EntryDatetime string
	Duration      int64
	Properties    []Property
}

func NewEntry(uid int64, project string, note string, entryDatetime string) Entry {
	var e Entry = Entry{uid, project, note, entryDatetime, 0, make([]Property, 0)}
	return e
}

func (e *Entry) AddEntryProperty(name string, value string) {
	var found bool = false
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, name) && strings.EqualFold(element.Value, value) {
			found = true
			break
		}
	}

	if !found {
		var property Property = NewProperty(e.Uid, name, value)
		e.Properties = append(e.Properties, property)
	}
}

func (e *Entry) GetTasksAsString() string {
	var result string

	// Count the number of Tasks.
	var taskCount = 0
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.TASK) {
			taskCount += 1
		}
	}

	// Append any Tasks to the string.
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.TASK) {
			result += element.Value
		}

		// Count backwards to add our separator.
		if taskCount > 1 {
			result += ", "
			taskCount -= 1
		}
	}

	return result
}

func (e *Entry) Dump() string {
	var result string

	if strings.EqualFold(e.Project, constants.BREAK) {
		result = "Break Time"
	} else {
		// Add the project.
		result = "Project[" + e.Project + "]"

		// Add the task(s).
		result = result + " Task[" + e.GetTasksAsString() + "]"
	}

	// Add the Date.
	result += " Date[" + e.EntryDatetime + "]"

	return result
}
