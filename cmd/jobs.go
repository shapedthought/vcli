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

var (folder string
	 customTemplate string
)

var jobsCmd = &cobra.Command{
	Use:   "job",
	Short: "Job",
	Long: `Job related commands.

ONLY WORKS WITH VBR AT THE MOMENT.

Subcommands:

Create Templates
vcli job template <job_id>

Create job from job file
vcli create job abc-job.yaml

Create jobs from folder
vcli create job -f .\path\to\jobs-folder

	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Template
		if args[0] == "template" {
			getTemplates(args, folder)
		} else if args[0] == "create" {
			createJob(args, folder, customTemplate)
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
	utils.SaveData(&varJobs, fmt.Sprintf("job-template-%s", varJobs.Name))

	// save storage data 
	utils.SaveData(&varJobs.Storage, fmt.Sprintf("storage-%s", varJobs.Name))

	// save guest processing data
	utils.SaveData(&varJobs.GuestProcessing, fmt.Sprintf("guest-processing-%s", varJobs.Name))

	// save schedule data
	utils.SaveData(&varJobs.Schedule, fmt.Sprintf("schedule-%s", varJobs.Name))

	// VirtualMachines.Includes already has correct type after model fix
	varVms := models.VirtualMachines{}
	varVms.Includes = varJobs.VirtualMachines.Includes
	varVms.Excludes = varJobs.VirtualMachines.Excludes

	var saveJob models.VbrJob

	saveJob.Name = varJobs.Name
	saveJob.Description = varJobs.Description
	saveJob.Type = varJobs.Type
	saveJob.IsDisabled = varJobs.IsDisabled
	saveJob.VirtualMachines = varVms

	jobName := fmt.Sprintf("job-%s", varJobs.Name)
	// save job data
	utils.SaveData(&saveJob, jobName) 

	fmt.Println("Job templates created.")

}

func getSettingsPath(cp string) string {
	settingsPath := os.Getenv("VCLI_SETTINGS_PATH")

	var sf string
	
	if len(cp) > 0 {
		sf = cp
	} else {
		sf = "job-template.yaml"
	}

	if len(settingsPath) != 0 {
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(settingsPath, "\\") {
				settingsPath = settingsPath + "\\" + sf
			} else {
				settingsPath = settingsPath + sf
			}
		} else {
			if !strings.HasSuffix(settingsPath, "/") {
				settingsPath = settingsPath + "/" + sf
			} else {
				settingsPath = settingsPath + sf
			}
		}
	} else {
		log.Fatal("VCLI_SETTINGS_PATH not set")
	}

	return settingsPath
}


func createJob(args []string, folder string, customTemplate string) {
	profile := utils.GetProfile()
	settings := utils.ReadSettings()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with vbr at the moment.")
	}

	var templateFile models.VbrJobPost

	settingsPath := getSettingsPath(customTemplate)

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

		// get all files in folder
		fmt.Println("Reading folder: ", folder)
		files, err := os.ReadDir(folder)
		if err != nil {
			log.Fatal("Error reading folder")
		}

		if len(files) == 0 {
			log.Fatal("No files found in folder")
		}

		var vbrJob models.VbrJob 

		for _, file := range files {
			fp := fmt.Sprintf("%s%s", folder, file.Name())
			if strings.Contains(fp, "job") {
				getYaml, err := os.Open(fp)
				if err != nil {
					log.Fatalf("Error opening job file: %v", err)
				}

				b, err := io.ReadAll(getYaml)
				utils.IsErr(err)

				err = yaml.Unmarshal(b, &vbrJob)
				utils.IsErr(err)

				templateFile.Type = vbrJob.Type
				templateFile.Name = vbrJob.Name
				templateFile.Description = vbrJob.Description
				templateFile.VirtualMachines = vbrJob.VirtualMachines
			}
			
			if strings.Contains(fp, "storage") {
				fmt.Println("Storage file found: ", file.Name())
				var vbrStorage models.Storage

				getYaml, err := os.Open(fp)
				if err != nil {
					log.Fatalf("Error opening storage file: %v", err)
				}

				b, err := io.ReadAll(getYaml)
				utils.IsErr(err)

				err = yaml.Unmarshal(b, &vbrStorage)
				utils.IsErr(err)

				templateFile.Storage = vbrStorage
			} 
			
			if strings.Contains(fp, "guest-processing") {
				fmt.Println("Guest processing file found: ", file.Name())
				var vbrGuestProcessing models.GuestProcessing

				getYaml, err := os.Open(fp)
				if err != nil {
					log.Fatalf("Error opening guest-processing file: %v", err)
				}

				b, err := io.ReadAll(getYaml)
				utils.IsErr(err)

				err = yaml.Unmarshal(b, &vbrGuestProcessing)
				utils.IsErr(err)

				templateFile.GuestProcessing = vbrGuestProcessing
			} 
			
			if strings.Contains(fp, "schedule") {
				fmt.Println("Schedule file found: ", file.Name())
				var vbrSchedule models.Schedule

				getYaml, err := os.Open(fp)
				if err != nil {
					log.Fatalf("Error opening schedule file: %v", err)
				}

				b, err := io.ReadAll(getYaml)
				utils.IsErr(err)

				err = yaml.Unmarshal(b, &vbrSchedule)
				utils.IsErr(err)

				templateFile.Schedule = vbrSchedule
			}
		}
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

	} else {
		fmt.Printf("No file or folder passed")
	}

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

}

func init() {
	jobsCmd.Flags().StringVarP(&folder, "folder", "f", "", "folder input")
	jobsCmd.Flags().StringVarP(&customTemplate, "template", "t", "", "custom template")
	rootCmd.AddCommand(jobsCmd)
}