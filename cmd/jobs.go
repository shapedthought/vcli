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
	"github.com/shapedthought/vcli/state"
	"github.com/shapedthought/vcli/utils"
	"github.com/shapedthought/vcli/vhttp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	folder         string
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

var (
	diffAll bool
)

var diffCmd = &cobra.Command{
	Use:   "diff [job-name]",
	Short: "Detect configuration drift from applied state",
	Long: `Compare current VBR job configuration against the last applied state
to detect manual changes or drift.

Each drift is classified by severity: CRITICAL, WARNING, or INFO.

Examples:
  # Check single job for drift
  vcli job diff SQL-Backup-Job

  # Check all jobs
  vcli job diff --all

  # Show only security-relevant drifts
  vcli job diff --all --security-only

  # Show only critical drifts
  vcli job diff --all --severity critical

Exit Codes:
  0 - No drift detected
  3 - Drift detected (INFO or WARNING)
  4 - Critical security drift detected
  1 - Error occurred`,
	Run: func(cmd *cobra.Command, args []string) {
		if diffAll {
			diffAllJobs()
		} else if len(args) > 0 {
			diffSingleJob(args[0])
		} else {
			log.Fatal("Provide job name or use --all")
		}
	},
}

func diffSingleJob(jobName string) {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	// Load from state
	stateMgr := state.NewManager()
	resource, err := stateMgr.GetResource(jobName)
	if err != nil {
		log.Fatalf("Job '%s' not found in state. Has it been applied?\n", jobName)
	}

	fmt.Printf("Checking drift for job: %s\n\n", jobName)

	// Fetch current from VBR
	endpoint := fmt.Sprintf("jobs/%s", resource.ID)
	current := vhttp.GetData[models.VbrJobGet](endpoint, profile)

	// Convert current job to map for comparison
	currentBytes, err := json.Marshal(current)
	if err != nil {
		log.Fatalf("Failed to marshal current job for drift detection: %v", err)
	}
	var currentMap map[string]interface{}
	if err := json.Unmarshal(currentBytes, &currentMap); err != nil {
		log.Fatalf("Failed to unmarshal current job into map for drift detection: %v", err)
	}

	// Compare, classify, enhance, filter
	drifts := detectDrift(resource.Spec, currentMap, jobIgnoreFields)
	drifts = classifyDrifts(drifts, jobSeverityMap)
	drifts = enhanceJobDriftSeverity(drifts)
	drifts = checkRepoHardeningDrift(drifts, resource.Spec)
	minSev := parseSeverityFlag()
	drifts = filterDriftsBySeverity(drifts, minSev)

	if len(drifts) == 0 {
		fmt.Println(noDriftMessage("Job matches applied state.", minSev))
		os.Exit(0)
	}

	// Display drift
	printSecuritySummary(drifts)
	fmt.Println("Drift detected:")
	for _, drift := range drifts {
		printDriftWithSeverity(drift)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  - %d drifts detected\n", len(drifts))
	fmt.Printf("  - Highest severity: %s\n", getMaxSeverity(drifts))
	fmt.Printf("  - Last applied: %s\n", resource.LastApplied.Format("2006-01-02 15:04:05"))
	fmt.Printf("  - Last applied by: %s\n", resource.LastAppliedBy)

	fmt.Println("\nThe job has drifted from the applied configuration.")
	fmt.Printf("\nTo reapply the desired state, run:\n")
	fmt.Printf("  vcli job apply <your-job-file>.yaml\n")

	os.Exit(exitCodeForDrifts(drifts))
}

func diffAllJobs() {
	loadSeverityOverrides()
	profile := utils.GetProfile()

	if profile.Name != "vbr" {
		log.Fatal("This command only works with VBR at the moment.")
	}

	stateMgr := state.NewManager()
	resources, err := stateMgr.ListResources("VBRJob")
	if err != nil {
		log.Fatalf("Failed to load state: %v\n", err)
	}

	if len(resources) == 0 {
		fmt.Println("No jobs in state.")
		return
	}

	fmt.Printf("Checking %d jobs for drift...\n\n", len(resources))

	minSev := parseSeverityFlag()
	driftedCount := 0
	cleanCount := 0
	var allDrifts []Drift

	for _, resource := range resources {
		// Fetch current from VBR
		endpoint := fmt.Sprintf("jobs/%s", resource.ID)
		current := vhttp.GetData[models.VbrJobGet](endpoint, profile)

		// Convert to map for comparison
		currentBytes, err := json.Marshal(current)
		if err != nil {
			fmt.Printf("  %s: Failed to marshal job data: %v\n", resource.Name, err)
			continue
		}
		var currentMap map[string]interface{}
		if err := json.Unmarshal(currentBytes, &currentMap); err != nil {
			fmt.Printf("  %s: Failed to unmarshal job data: %v\n", resource.Name, err)
			continue
		}

		// Detect, classify, enhance, filter
		drifts := detectDrift(resource.Spec, currentMap, jobIgnoreFields)
		drifts = classifyDrifts(drifts, jobSeverityMap)
		drifts = enhanceJobDriftSeverity(drifts)
		drifts = checkRepoHardeningDrift(drifts, resource.Spec)
		drifts = filterDriftsBySeverity(drifts, minSev)

		if len(drifts) > 0 {
			maxSev := getMaxSeverity(drifts)
			fmt.Printf("  %s %s: %d drifts detected\n", maxSev, resource.Name, len(drifts))
			allDrifts = append(allDrifts, drifts...)
			driftedCount++
		} else {
			fmt.Printf("  %s: No drift\n", resource.Name)
			cleanCount++
		}
	}

	if driftedCount > 0 {
		fmt.Println()
		printSecuritySummary(allDrifts)
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("  - %d jobs clean\n", cleanCount)
	fmt.Printf("  - %d jobs drifted\n", driftedCount)

	if driftedCount > 0 {
		os.Exit(exitCodeForDrifts(allDrifts))
	}
	os.Exit(0)
}

func init() {
	diffCmd.Flags().BoolVar(&diffAll, "all", false, "Check drift for all jobs in state")
	addSeverityFlags(diffCmd)
	jobsCmd.AddCommand(diffCmd)

	jobsCmd.Flags().StringVarP(&folder, "folder", "f", "", "folder input")
	jobsCmd.Flags().StringVarP(&customTemplate, "template", "t", "", "custom template")
	rootCmd.AddCommand(jobsCmd)
}