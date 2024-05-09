/*
Copyright © 2023 Jeff Lanzarotta
*/
package cmd

import (
	"log"
	"os"
	"timetracker/constants"

	"github.com/golang-module/carbon/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"timetracker/internal/database"
	"timetracker/internal/models"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show various information",
	Run: func(cmd *cobra.Command, args []string) {
		runShow(cmd, args)
	},
}

var favorites bool
var statistics bool

type Configuration struct {
	DatabaseFilename string     `yaml:"database_file"`
	WeekStart        string     `yaml:"week_start"`
	RoundToMinutes   int        `yaml:"round_to_minutes"`
	Debug            bool       `yaml:"debug"`
	Favorites        []Favorite `yaml:"favorites"`
}

type Favorite struct {
	Favorite string `yaml:"favorite"`
}

func init() {
	showCmd.Flags().BoolVarP(&favorites, constants.FAVORITES, constants.EMPTY, false, "Show favorites")
	showCmd.Flags().BoolVarP(&statistics, constants.STATISTICS, constants.EMPTY, false, "Show statistics")
	rootCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runShow(cmd *cobra.Command, _ []string) {
	// Get the --favorites flag.
	favorites, _ := cmd.Flags().GetBool(constants.FAVORITES)
	statistics, _ := cmd.Flags().GetBool(constants.STATISTICS)

	if favorites {
		showFavorites()
	}

	if statistics {
		showStatistics()
	}
}

func showFavorites() {
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

	log.Printf("Favorites found in configuration file[%s]:\n", viper.ConfigFileUsed())

	for i, f := range config.Favorites {
		log.Printf("Favorite %d: [%s]\n", i, f.Favorite)
	}
}

func showStatistics() {
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var firstEntry models.Entry = db.GetFirstEntry()
	var lastEntry models.Entry = db.GetLastEntry()
	var count int64 = db.GetCountEntries()

	log.Printf("\n")

	var lastDateTime carbon.Carbon = carbon.Parse(lastEntry.EntryDatetime)
	var firstDateTime carbon.Carbon = carbon.Parse(firstEntry.EntryDatetime)
	var diff int64 = firstDateTime.DiffInSeconds(lastDateTime)

	var t table.Writer = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.AppendHeader(table.Row{"Statistic", "Value"})
	t.AppendRow(table.Row{"First Entry", firstEntry.Dump(false)})
	t.AppendRow(table.Row{"Last Entry", lastEntry.Dump(false)})
	t.AppendRow(table.Row{"Total Records", count})
	t.AppendRow(table.Row{"Total Duration", secondsToHuman(diff)})
	log.Println(t.Render())
}
