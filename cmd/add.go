package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
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
completed task to the database with an optional note.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAdd(cmd, args)
	},
}

var favorite int

func getFavorite(index int) Favorite {
	if index < 0 {
		log.Fatalf("Favorite must be >= 0.")
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

	if index >= len(config.Favorites) {
		log.Fatalf("Favorite[%d] not found in configuration file[%s].\n", index, viper.ConfigFileUsed())
		os.Exit(1)
	}

	return config.Favorites[index]
}

func init() {
	// Here you will define your flags and configuration settings.
	addCmd.Flags().StringVarP(&at, constants.AT, constants.EMPTY, constants.EMPTY, constants.NATURAL_LANGUAGE_DESCRIPTION)
	addCmd.Flags().StringVarP(&note, constants.NOTE, constants.EMPTY, constants.EMPTY, constants.NOTE_DESCRIPTION)
	addCmd.Flags().IntVarP(&favorite, constants.FAVORITE, constants.EMPTY, -1, "Favorite")
	addCmd.Flags().Lookup(constants.FAVORITE).NoOptDefVal = "-1"
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {
	// Get the current date/time.
	var addTime carbon.Carbon = carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString(constants.AT)

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
	var url string = constants.EMPTY

	// If a project+task was passed in, use that project+task combination.  If it was not, see if a
	// favorite was passed in.
	if len(args) > 0 {
		projectTask = args[0]
	} else {
		// Get the --favorite flag.
		favorite, err := cmd.Flags().GetInt(constants.FAVORITE)
		if err != nil {
			log.Fatalf("Fatal: Missing project+task or --favorite.")
			os.Exit(1)
		} else {
			if favorite < 0 {
				for {
					showFavorites()

					// Prompt the user for the index number of the filename they would like to send.
					r := bufio.NewReader(os.Stdin)

					fmt.Fprintf(os.Stderr, "\nPlease enter the number of the favorite to add; otherwise, [Return] to quit. > ")
					var s, _ = r.ReadString('\n')
					s = strings.TrimSpace(s)

					// If the result is empty, use the original passed in value.
					if len(s) <= 0 {
						os.Exit(0)
					}

					// Convert the string to an integer, thus validating the user entered a number.
					i, err := strconv.Atoi(s)
					if err != nil {
						log.Printf("Invalid number entered.\n")
						continue
					}

					var fav Favorite = getFavorite(i)
					projectTask = fav.Favorite
					url = fav.URL
					break
				}
			} else {
				var fav Favorite = getFavorite(favorite)
				projectTask = fav.Favorite
				url = fav.URL
			}
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

	// If a URL was configured for this project+task, add it to the entry.
	if len(url) > 0 {
		entry.AddEntryProperty(constants.URL, url)
	}

	log.Printf("Adding %s.\n", entry.Dump(false))

	// Write the new Entry to the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	db.InsertNewEntry(entry)
}
