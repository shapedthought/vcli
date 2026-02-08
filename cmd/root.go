/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/shapedthought/owlctl/config"
	"github.com/spf13/cobra"
)

var targetFlag string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "owlctl",
	Short: "A CLI application for Veeam APIs",
	Long:  `A CLI application that works with all Veeam APIs`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if targetFlag == "" {
			return nil
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load owlctl.yaml: %w", err)
		}

		target, err := cfg.GetTarget(targetFlag)
		if err != nil {
			return fmt.Errorf("--target %q: %w", targetFlag, err)
		}

		if err := os.Setenv("OWLCTL_URL", target.URL); err != nil {
			return fmt.Errorf("failed to set OWLCTL_URL: %w", err)
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVar(&targetFlag, "target", "", "Named VBR target from owlctl.yaml")
}
