package cmd

import (
	"log"
	"os"
	"timetracker/constants"

	"github.com/spf13/cobra"
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
	purgeCmd.Flags().BoolP(constants.PREVIOUS_YEAR, constants.EMPTY, false, "Purge the previous year's entries.")
	rootCmd.AddCommand(purgeCmd)
}

func runPurge(cmd *cobra.Command, args []string) {

	all, _ := cmd.Flags().GetBool(constants.ALL)
	previousYear, _ := cmd.Flags().GetBool(constants.PREVIOUS_YEAR)

	if all {
		yesNo := yesNoPrompt("Are you sure you want to purge ALL the entries from your database?")
		if yesNo {
			reallySure := yesNoPrompt("Are you REALLY sure you want to purge ALL the entries from your database?")
			if reallySure {
				// Yes was enter, so purge ALL entries.
				log.Printf("All entries purged.\n")
			}
		} else {
			log.Printf("Nothing purged.\n")
		}
	}

	if previousYear {

	}

	log.Fatalf("Not implemented... yet.")
	os.Exit(1)
}