/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "owlctl",
	Short: "A CLI application for Veeam APIs",
	Long:  `A CLI application that works with all Veeam APIs`,

}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
