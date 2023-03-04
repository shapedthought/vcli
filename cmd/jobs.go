package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var folder string

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Jobs",
	Long: `Jobs related commands.

ONLY WORKS WITH VBR AT THE MOMENT.

Subcommands:

Create Templates
vcli jobs template <job_id>

Create job from job file
vcli create jobs abc-job.yaml

Create jobs from folder
vcli create jobs -f /path/to/jobs-folder

	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Template
		if args[0] == "template" {
			getTemplates(args, folder)
		} else if args[0] == "create" {
			createJob(args, folder)
		}
	},
}

func getTemplates(args []string, folder string) {

	// get the job data
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with vbr at the moment.")
	}

	ju := fmt.Sprintf("jobs/%s", args[1])

	varJobs := vhttp.GetData[models.VbrJobGet](ju, profile)

	// save base job file
	utils.SaveData(&varJobs, "job-template")

	// save storage data
	utils.SaveData(&varJobs.Storage, "storage")

	// save guest processing data
	utils.SaveData(&varJobs.GuestProcessing, "guest-processing")

	// save schedule data
	utils.SaveData(&varJobs.Schedule, "schedule")

	var inc []models.Includes

	for _, v := range varJobs.VirtualMachines.Includes {
		ti := models.Includes{
			Type: v.InventoryObject.Type,
			HostName: v.InventoryObject.HostName,
			Name: v.InventoryObject.Name,
			ObjectID: v.InventoryObject.ObjectID,
		}

		inc = append(inc, ti)
	}

	varVms := models.VirtualMachines{}
	varVms.Includes = inc
	varVms.Excludes = varJobs.VirtualMachines.Excludes

	var saveJob models.VbrJob

	saveJob.Name = varJobs.Name
	saveJob.Description = varJobs.Description
	saveJob.Type = varJobs.Type
	saveJob.IsDisabled = varJobs.IsDisabled
	saveJob.VirtualMachines = varVms

	jobName := fmt.Sprintf("%s-job", varJobs.Name)
	// save job data
	utils.SaveData(&saveJob, jobName) 

}

func createJob(args []string, folder string) {
	profile := utils.GetProfile()
	settings := utils.ReadSettings()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with vbr at the moment.")
	}

	var templateFile models.VbrJobPost

	settingsPath := os.Getenv("VCLI_SETTINGS_PATH")
	if len(settingsPath) != 0 {
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(settingsPath, "\\") {
				settingsPath = settingsPath + "\\" + "job-template.yaml"
			} else {
				settingsPath = settingsPath + "job-template.yaml"
			}
		} else {
			if !strings.HasSuffix(settingsPath, "/") {
				settingsPath = settingsPath + "/" + "job-template.yaml"
			} else {
				settingsPath = settingsPath + "job-template.yaml"
			}
		}
	} else {
		log.Fatal("VCLI_SETTINGS_PATH not set")
	}

	fmt.Println(settingsPath)
	getYaml, err := os.Open(settingsPath)
	if err != nil {
		log.Fatal("Error opening job-template.yaml")
	}

	b, err := io.ReadAll(getYaml)
	utils.IsErr(err)

	err = yaml.Unmarshal(b, &templateFile)
	utils.IsErr(err)

	api_url := utils.GetAddress(profile, settings)

	if len(folder) > 0  {
		fmt.Printf("Folder, flag: %v", folder)
	} else if len(args) > 1 {

		var vbrJob models.VbrJob 

		getYaml, err := os.Open(args[1])
		
		utils.IsErr(err)

		b, err := io.ReadAll(getYaml)
		utils.IsErr(err)

		err = yaml.Unmarshal(b, &vbrJob)
		utils.IsErr(err)

		utils.SaveJson(&templateFile, fmt.Sprintf("%s-job", vbrJob.Name))

		templateFile.Type = vbrJob.Type
		templateFile.Name = vbrJob.Name
		templateFile.Description = vbrJob.Description
		templateFile.VirtualMachines = vbrJob.VirtualMachines

		utils.SaveJson(&templateFile, fmt.Sprintf("%s-job", vbrJob.Name))

		sendData, err := json.Marshal(templateFile)
		utils.IsErr(err)

		connstring := fmt.Sprintf("https://%v:%v%v%v/%v", api_url, profile.Port, "/api/", profile.APIVersion, "/jobs")

		r, err := http.NewRequest("POST", connstring, bytes.NewReader(sendData))

		utils.IsErr(err)
		headers := utils.ReadHeader[models.SendHeader]()
		r.Header.Add("x-api-version", profile.Headers.XAPIVersion)
		r.Header.Add("Content-Type", "application/json")

		r.Header.Add("Authorization", "Bearer "+headers.AccessToken)

		client := vhttp.Client(settings.ApiNotSecure)
		res, err := client.Do(r)
		if err != nil {
			fmt.Printf("Error sending HTTP request %v\n", err)
			return
		}

		defer res.Body.Close()

		fmt.Println("Status Code:", res.StatusCode)
		fmt.Println("Status:", res.Status)
	} else {
		fmt.Printf("No file or folder passed")
	}

}

func init() {
	jobsCmd.Flags().StringVarP(&folder, "folder", "f", "", "folder input")
	rootCmd.AddCommand(jobsCmd)
}