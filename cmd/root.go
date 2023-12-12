/*
Copyright © 2023 Jeff Lanzarotta
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"errors"
	"log"
	"os"
	"timetracker/constants"
	"timetracker/internal/database"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// var BuildDateTime string
//var version = "1.0.1.0"

//var version string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "timetracker",
	Short: "Simple program used to track time spent on projects and tasks.",
	Long: `Time Tracker is a simple command line tool use to track the time you spend
on a specific project and the one or more tasks associated with that project.
It was inspired by the concepts of utt (Ultimate Time Tracker) and timetrap.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", constants.EMPTY, "config file (default is $HOME/.timetracker.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "Noninteractive, assume yes as answer to all prompts")
	rootCmd.PersistentFlags().Bool("debug", false, "Display stack traces for errors")
	rootCmd.PersistentFlags().Bool("help", false, "Show help for command")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".timetracker" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".timetracker.yaml")
	}

	// Read in environment variables that match.
	viper.AutomaticEnv()

	// Set default database.
	viper.SetDefault("database_file", os.ExpandEnv("$HOME/.timetracker.db"))

	// Round to 15 minute intervals by default.
	viper.SetDefault("round_to_minutes", 15)

	// Require a note.
	viper.SetDefault("require_note", false)

	// Set day of the week when determining start of the week.
	viper.SetDefault("week_start", "Monday")

	// Set debug to false.
	viper.SetDefault("debug", false)

	// Read the configuration file.
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// No config file, just use defaults.
			log.Println("unable to load config file, using default values.")
		} else {
			log.Fatalf("Fatal error reading config file: %s\n", err.Error())
			os.Exit(1)
		}
	}

	// Dump our some debug information.
	if viper.GetBool("debug") {
		log.Printf("%s = [%s]\n", constants.WEEK_START, viper.GetString(constants.WEEK_START))
		log.Printf("%s = [%d]\n", constants.ROUND_TO_MINUTES, viper.GetInt64(constants.ROUND_TO_MINUTES))
	}

	// Check if the database exists or not.  If it does not, create it.
	_, err = os.Stat(viper.GetString(constants.DATABASE_FILE))
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("Database[%s] does not exist, creating...", viper.GetString(constants.DATABASE_FILE))

		var filename string = viper.GetString(constants.DATABASE_FILE)
		os.Create(filename)

		db := database.New(viper.GetString(constants.DATABASE_FILE))
		db.Create()
	}
}
