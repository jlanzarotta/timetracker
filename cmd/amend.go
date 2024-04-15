/*
Copyright © 2024 Jeff Lanzarotta
*/
package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"timetracker/constants"

	"github.com/golang-module/carbon/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"timetracker/internal/database"
	"timetracker/internal/models"
)

// amendCmd represents the amend command
var amendCmd = &cobra.Command{
	Use: "amend",
	//Args:  cobra.ExactArgs(0),
	Args:  cobra.MaximumNArgs(1),
	Short: "Amend an entry",
	Long: `Amend is a convenient way to modify an entry, default is the last
entry.  It lets you modify the project, task, and/or datetime.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAmend(cmd, args)
	},
}

func init() {
	amendCmd.Flags().BoolP("today", constants.EMPTY, false, "List all the entries for today.")
	rootCmd.AddCommand(amendCmd)
}

func runAmend(cmd *cobra.Command, args []string) {
	var entry models.Entry

	today, _ := cmd.Flags().GetBool("today")
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	if today {
		var input_value string = constants.EMPTY

		for {
			var t table.Writer = table.NewWriter()
			t.SetAutoIndex(true)
			t.AppendHeader(table.Row{"Project", "Task(s)", "Date/Time"})
			var entries []models.Entry = db.GetEntriesForToday(carbon.Now().StartOfDay(), carbon.Now().EndOfDay())
			for _, entry := range entries {
				t.AppendRow(table.Row{entry.Project, entry.GetTasksAsString(), entry.EntryDatetime})
			}

			fmt.Println(t.Render())

			fmt.Print("Please enter index number of the entry you would like to amend; otherwise, ENTER to quit...\n")
			n, _ := fmt.Scanln(&input_value)

			// If nothing was entered, break out of the loop.
			if n <= 0 {
				log.Printf("No entry amended.\n")
				return
			}

			// Validate what the user entered is actually a number.
			i, err := strconv.Atoi(input_value)
			if err != nil {
				fmt.Printf("\nPlease enter a valid value.\n\n")
				continue
			}

			// Validate that the entry was between 1 and the length of the entries.
			if i <= 0 || i > len(entries) {
				fmt.Printf("\nPlease enter a valid value.\n\n")
				continue
			}

			// Get the entry the user wants to amend.
			entry = entries[i - 1]
			break
		}
	} else {
		// Get the last Entry from the database.
		entry = db.GetLastEntry()
	}

	log.Printf("Amending...\n" + entry.Dump(true) + "\n\n")

	// Prompt to change project.
	newProject := prompt("Project", entry.Project)
	newTask := prompt("Task", entry.GetTasksAsString())
	newNote := prompt("Note", entry.Note)
	newEntryDatetime := prompt("Date Time", entry.EntryDatetime)

	// Validate that the user entered a correctly formatted date/time.
	e := carbon.Parse(newEntryDatetime)
	if e.Error != nil {
		log.Fatalf("Invalid ISO8601 date/time format.  Please try to amend again with a valid ISO8601 formatted date/time.")
	} else {
		newEntryDatetime = carbon.Parse(newEntryDatetime).ToIso8601String()
	}

	log.Printf("\n")

	// Create a table to show the old verses new values.
	var t table.Writer = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.AppendHeader(table.Row{"", "Old", "New"})
	t.AppendRow(table.Row{"Project", entry.Project, newProject})
	t.AppendRow(table.Row{"Task", entry.GetTasksAsString(), newTask})
	t.AppendRow(table.Row{"Note", entry.Note, newNote})
	t.AppendRow(table.Row{"Datetime", entry.EntryDatetime, newEntryDatetime})

	// Render the table.
	fmt.Println(t.Render())

	// Ask the user if they want to commit these changes or not.
	yesNo := yesNoPrompt("\nCommit these changes?")
	if yesNo {
		var e database.Entry
		e.Uid = entry.Uid
		e.Project = newProject
		e.Note = sql.NullString{String: newNote, Valid: true}
		e.Name = sql.NullString{String: constants.TASK, Valid: true}
		e.Value = sql.NullString{String: newTask, Valid: true}
		e.EntryDatetime = newEntryDatetime
		db.UpdateEntry(e)

		log.Printf("Last entry amended.\n")
	} else {
		log.Printf("Last entry not amended.\n")
	}
}

func prompt(label string, value string) string {
	r := bufio.NewReader(os.Stdin)
	var s string

	fmt.Fprintf(os.Stderr, "Enter %s (empty for no change) ["+value+"] : ", label)
	s, _ = r.ReadString('\n')
	s = strings.TrimSpace(s)

	// If the result is empty, use the original passed in value.
	if len(s) <= 0 {
		s = value
	}

	return s
}
