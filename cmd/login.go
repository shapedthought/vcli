/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com
*/
package cmd

import (
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
)

var logout bool

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Logs into the API",
	Long:  `Logs into the API set in the active profile`,
	Run: func(cmd *cobra.Command, args []string) {
		if !logout {
			vhttp.ApiLogin()
		} else {
			vhttp.ApiLogout()
		}
	},
}

func init() {
	loginCmd.Flags().BoolVarP(&logout, "logout", "l", false, "logs out of the API")
	rootCmd.AddCommand(loginCmd)
}
