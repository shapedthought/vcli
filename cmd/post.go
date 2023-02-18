package cmd

import (
	"log"
	"os"

	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
)

// var usi bool

// postCmd represents the post command
var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Sends a POST command to API",
	Long: `Sends a POST commands to the selected profile.

Payload needs to be in the JSON format.

Note that vcli does not type check the payload.

Commands:
vcli post jobs/c69eb538-5a07-4bd7-80cb-bdf5142eadd6/start
vcli post jobs -f job.json

	`,
	Run: func(cmd *cobra.Command, args []string) {

		profile := utils.GetProfile()
		settings := utils.ReadSettings()

		var api_url string

		if settings.CredsFileMode {
			if len(profile.Address) > 0 {
				api_url = profile.Address
			} else {
				log.Fatal("Profile Address not set")
			}
		} else {
			api_url = os.Getenv("VCLI_URL")
			if api_url == "" {
				log.Fatal("VCLI_URL environment variable not set")
			}
		}

		vhttp.SendData(api_url, filename, args[0], "POST", profile, settings)

	},
}

func init() {
	postCmd.Flags().StringVarP(&filename, "file", "f", "", "payload in json format")
	// postCmd.Flags().BoolVarP(&usi, "stdin", "s", false, "read payload from stdin")
	// postCmd.MarkFlagRequired("filename")
	rootCmd.AddCommand(postCmd)

}
