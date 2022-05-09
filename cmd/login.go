/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com

*/
package cmd

import (
	"fmt"

	"github.com/shapedthought/veeamcli/vhttp"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Logs into the API",
	Long:  `Logs into the API set in the active profile`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("login called")
		logInApi()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func logInApi() {
	vhttp.ApiLogin()
}
