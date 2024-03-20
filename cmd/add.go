/*
Copyright © 2023 Jeff Lanzarotta
*/
package cmd

import (
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
	"gopkg.in/yaml.v3"

	"timetracker/internal/database"
	"timetracker/internal/models"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add project+task...",
	Args:  cobra.MaximumNArgs(1),
	Short: "Add a completed task",
	Long: `Once you have completed a task, use this command to add that newly
completed task to the database with an optional note`,
	Run: func(cmd *cobra.Command, args []string) {
		runAdd(cmd, args)
	},
}

var favorite int

func getFavorite(index int) string {
	if index < 0 {
		log.Fatalf("Fatal: Favorite must be >= 0.")
		os.Exit(1)
	}

	data, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		log.Fatalf("Error reading configuration file[%s]. %s\n", viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	var config Configuration

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling configuration file[%s]. %s\n", viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	return config.Favorites[index].Favorite
}

func init() {
	addCmd.Flags().StringVarP(&at, "at", constants.EMPTY, constants.EMPTY, constants.NATURAL_LANGUAGE_DESCRIPTION)
	addCmd.Flags().StringVarP(&note, "note", constants.EMPTY, constants.EMPTY, constants.NOTE_DESCRIPTION)
	addCmd.Flags().IntVarP(&favorite, "favorite", constants.EMPTY, -1, "Favorite")
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runAdd(cmd *cobra.Command, args []string) {
	// Get the current date/time.
	var addTime carbon.Carbon = carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString("at")

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("Fatal parsing 'at' time. %s\n", err.Error())
			os.Exit(1)
		}

		addTime = carbon.CreateFromStdTime(atTime)
	}

	var projectTask string = constants.EMPTY

	// If a project+tast was passed in, use that project+task combination.  If it was not, see if a
	// favorite was passed in.
	if len(args) > 0 {
		projectTask = args[0]
	} else {
		// Get the --favorite flag.
		favorite, err := cmd.Flags().GetInt("favorite")
		if err != nil {
			log.Fatalf("Fatal: Missing project+task or --favorite.")
			os.Exit(1)
		} else {
			projectTask = getFavorite(favorite)
		}
	}

	// Split the project/task into pieces.
	var pieces []string = strings.Split(projectTask, constants.TASK_DELIMITER)
	if len(pieces) < 2 {
		log.Fatalf("Fatal parsing 'project+task'.  Malformed project+task.\n")
		os.Exit(1)
	}

	// Create a new Entry.
	var entry models.Entry = models.NewEntry(constants.UNKNOWN_UID, pieces[0], note,
		addTime.ToRfc3339String())

	// Populate the newly created Entry with its tasks.
	for i := 1; i < len(pieces); i += 1 {
		entry.AddEntryProperty(constants.TASK, pieces[i])
	}

	log.Printf("Adding %s\n", entry.Dump())

	// Write the new Entry to the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	db.InsertNewEntry(entry)
}
