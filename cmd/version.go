package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

var BuildVersion string
// var vedrsion string = "1.0.1.0"
var BuildDateTime string

// addCmd represents the add command
var versionCmd = &cobra.Command{
	Use:    "version",
	Short:  "Show the version information",
	Long:   "Show the version information",
	Run: func(cmd *cobra.Command, args []string) {
		runVersion(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	log.Printf("  Version: " + BuildVersion + "\n" +
		"Copyright: (c) 2018-" + time.Now().Format("2006") +
		" Jeff Lanzarotta, All rights reserved\n  Born on: " + BuildDateTime + "\n")
}
