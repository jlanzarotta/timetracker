package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"timetracker/constants"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/golang-module/carbon/v2"
	"github.com/ijt/go-anytime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"timetracker/internal/database"
	"timetracker/internal/models"
)

// stretchCmd represents the add command
var stretchCmd = &cobra.Command{
	Use:   "stretch last project",
	Short: "Stretch the latest entry",
	Long:  "Stretch the latest entry to 'now' or the whatever is specified using the 'at' flag command",
	Run: func(cmd *cobra.Command, args []string) {
		runStretch(cmd, args)
	},
}

func init() {
	stretchCmd.Flags().StringVarP(&at, "at", constants.EMPTY, constants.EMPTY, "Natural Language Time, e.g., '18 minutes ago'")
	rootCmd.AddCommand(stretchCmd)
}

func runStretch(cmd *cobra.Command, _ []string) {
	// Get the current date/time.
	var stretchTime carbon.Carbon = carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString("at")

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("Fatal parsing 'at' time. %s\n", err.Error())
			os.Exit(1)
		}

		stretchTime = carbon.CreateFromStdTime(atTime)
	}

	// Get the last Entry from the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var entry models.Entry = db.GetLastEntry()

	// Create the prompt.
	var prompt string = "Would you like to stretch Project[" + entry.Project + "]"
	if len(entry.GetTasksAsString()) > 0 {
		prompt = prompt + " Tasks[" + entry.GetTasksAsString() + "]"
	}
	prompt = prompt + " to " + stretchTime.ToCookieString() + "?"

	// Ask the user if they actually want to stretch the last entry or not.
	yesNo := yesNoPrompt(prompt)
	if yesNo {
		// Yes was enter, so update the latest.
		var e database.Entry
		e.Uid = entry.Uid
		e.EntryDatetime = stretchTime.ToIso8601String()
		db.UpdateEntry(e)

		log.Printf("Last entry was stretched.\n")
	} else {
		log.Printf("Last entry was NOT stretched.\n")
	}
}

func yesNoPrompt(label string) bool {
	choices := "Y/N (yes/no)"

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		} else if s == "n" || s == "no" {
			return false
		} else {
			return false
		}
	}
}
