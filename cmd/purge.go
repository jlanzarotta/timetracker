package cmd

import (
	"fmt"
	"log"
	"os"
	"timetracker/constants"
	"timetracker/internal/database"

	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Args:  cobra.MinimumNArgs(0),
	Short: "Purge entries from the sqlite database",
	Long: `As you continuously add completed entries, the database continues to go unbounded.  The purge command allows you to mange the database size.`,
	Run: func(cmd *cobra.Command, args []string) {
		runPurge(cmd, args)
	},
}

func init() {
	purgeCmd.Flags().BoolP(constants.ALL, constants.EMPTY, false, "Purge ALL entries.  Use with extreme caution!!!")
	purgeCmd.Flags().BoolP(constants.PREVIOUS_YEARS, constants.EMPTY, false, "Purge the non-current year's entries.")
	rootCmd.AddCommand(purgeCmd)
}

func runPurge(cmd *cobra.Command, args []string) {
	all, _ := cmd.Flags().GetBool(constants.ALL)
	previousYears, _ := cmd.Flags().GetBool(constants.PREVIOUS_YEARS)

	if all {
		yesNo := yesNoPrompt("Are you sure you want to purge ALL the entries from your database?")
		if yesNo {
			reallySure := yesNoPrompt("WARNING: Are you REALLY sure you want to purge ALL the entries from your database?")
			if reallySure {
				reallyReallySure := yesNoPrompt("LAST WARNING: Are you REALLY REALLY sure you want to purge ALL the entries from your database?")
				if reallyReallySure {
					// Yes was enter, so purge ALL entries.
					db := database.New(viper.GetString(constants.DATABASE_FILE))
					db.PurgeAllEntries()
					log.Printf("All entries purged.\n")
				} else {
					log.Printf("Nothing purged.\n")
				}
			} else {
				log.Printf("Nothing purged.\n")
			}
		} else {
			log.Printf("Nothing purged.\n")
		}
	}

	if previousYears {
		var year int = carbon.Now().Year()
		var prompt = fmt.Sprintf("Are you sure you want to purge all entries prior to %d from the database?", year);
		yesNo := yesNoPrompt(prompt)
		if yesNo {
			db := database.New(viper.GetString(constants.DATABASE_FILE))
			db.PurgePreviousYearsEntries()
			log.Printf("All entries prior to %d have been purged.", year)
		} else {
			log.Printf("Nothing purged.\n")
		}
	}

	os.Exit(1)
}