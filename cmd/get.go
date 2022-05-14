/*
Copyright © 2022 Ed Howard exfhoward@protonmail.com

*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/cheynewallace/tabby"
	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/shapedthought/veeamcli/vbrmodels"
	"github.com/shapedthought/veeamcli/vhttp"
	"github.com/spf13/cobra"
)

var save bool

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets data from the API",
	Long: `Sends a GET request to a specified endpoint.

Prints a table of the specified resource. 

Examples:
veeamcli get jobStates
veeamcli get jobs 
veeamcli get proxies
veeamcli get repos
veeamcli get sessions - last 20
veeamcli get backupObjects
veeamcli get inventory - gets VMware inventory
`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := utils.GetProfile()

		if profile.Name == "vbr" {
			switch args[0] {
			case "jobStates":
				getVbrJobStates(profile)
			case "jobs":
				getVbrJobs(profile)
			case "proxies":
				getProxies(profile)
			case "repos":
				getRepos(profile)
			case "sobr":
				getSobr(profile)
			case "sessions":
				getSessions(profile)
			case "backupObjects":
				getBackupObject(profile)
			case "inventory":
				getInventory(profile)
			case "all":
				getAll(profile)
			default:
				fmt.Println("type not found")
			}
		}
	},
}

func init() {
	getCmd.Flags().BoolVarP(&save, "save", "s", false, "saves the data to a yaml file")
	rootCmd.AddCommand(getCmd)

}

func getVbrJobStates(profile models.Profile) {
	jobData := vhttp.GetData[vbrmodels.VbrJobStates]("jobs/states", profile)

	t := tabby.New()

	t.AddHeader("NAME", "TYPE", "STATUS", "DESCRIPTION", "REPOSITORY", "OBJECT COUNT")
	for _, i := range jobData.Data {
		t.AddLine(i.Name, i.Type, i.Status, i.Description, i.RepositoryName, strconv.Itoa(i.ObjectsCount))
	}

	fmt.Println("")
	t.Print()
	fmt.Println("")

	if save {
		utils.SaveData(&jobData, "vbrJobStates")
	}

}

func getVbrJobs(profile models.Profile) {
	// get the job states which have the name and id
	jobs := vhttp.GetData[vbrmodels.VbrJobs]("jobs", profile)

	repos := vhttp.GetData[vbrmodels.Repos]("backupInfrastructure/repositories", profile)

	t := tabby.New()
	t.AddHeader(
		"ENABLED",
		"NAME",
		"VMs",
		"APP AWARE",
		"PROXY AUTO",
		"REPO NAME",
		"MODE",
		"SYNTH",
		"ENCRYPT",
		"DAYS",
		"WEEKS",
		"MONTHS",
		"YEARS")

	for _, j := range jobs.Data {
		vms := len(j.VirtualMachines.Includes)
		gw := 0
		gm := 0
		gy := 0
		rn := repos.GetName(j.Storage.BackupRepositoryID)
		if j.Storage.GfsPolicy.IsEnabled {
			if j.Storage.GfsPolicy.Weekly.IsEnabled {
				gw = j.Storage.GfsPolicy.Weekly.KeepForNumberOfWeeks
			}
			if j.Storage.GfsPolicy.Monthly.IsEnabled {
				gm = j.Storage.GfsPolicy.Monthly.KeepForNumberOfMonths
			}
			if j.Storage.GfsPolicy.Yearly.IsEnabled {
				gy = j.Storage.GfsPolicy.Yearly.KeepForNumberOfYears
			}
		}

		t.AddLine(
			!j.IsDisabled,
			j.Name,
			vms,
			j.GuestProcessing.AppAwareProcessing.IsEnabled,
			j.Storage.BackupProxies.AutoSelection,
			rn,
			j.Storage.AdvancedSettings.BackupModeType,
			j.Storage.AdvancedSettings.SynthenticFulls.IsEnabled,
			j.Storage.AdvancedSettings.StorageData.Encryption.IsEnabled,
			j.Storage.RetentionPolicy.Quantity,
			gw,
			gm,
			gy,
		)
	}

	fmt.Println("")
	t.Print()
	fmt.Println("")

	if save {
		utils.SaveData(&jobs, "vbrJobs")
	}

}

func getProxies(profile models.Profile) {

	p := vhttp.GetData[vbrmodels.Proxies]("backupInfrastructure/proxies", profile)

	t := tabby.New()

	t.AddHeader("NAME", "DESCRIPTION", "TYPE", "MAX TASKS")

	for _, i := range p.Data {
		t.AddLine(i.Name, i.Description, i.Type, i.Server.MaxTaskCount)
	}

	fmt.Println("")
	t.Print()
	fmt.Println("")

	if save {
		utils.SaveData(&p, "proxies")
	}

}

func getRepos(profile models.Profile) {
	rst := vhttp.GetData[vbrmodels.RepoStates]("backupInfrastructure/repositories/states", profile)

	var rd []vbrmodels.Repo

	for _, r := range rst.Data {
		st := fmt.Sprintf("backupInfrastructure/repositories/%v", r.ID)
		d := vhttp.GetData[vbrmodels.Repo](st, profile)
		rd = append(rd, d)
	}

	t := tabby.New()

	t.AddHeader(
		"NAME",
		"HOST NAME",
		"PATH",
		"MAX TASKS",
		"PER VM",
		"CAP GB",
		"CAP USED GB")

	for _, r := range rst.Data {
		mx := 0
		pvm := false
		for _, rs := range rd {
			if r.ID == rs.ID {
				mx = rs.Repository.MaxTaskCount
				pvm = rs.Repository.AdvancedSettings.PerVMBackup
			}
		}

		t.AddLine(
			r.Name,
			r.HostName,
			r.Path,
			mx,
			pvm,
			r.CapacityGB,
			r.UsedSpaceGB,
		)
	}

	fmt.Println("")
	t.Print()
	fmt.Println("")

	if save {
		utils.SaveData(&rst, "repoStates")
	}

}

func getSobr(profile models.Profile) {
	sobr := vhttp.GetData[vbrmodels.Sobr]("backupInfrastructure/scaleOutRepositories", profile)

	t := tabby.New()

	t.AddHeader(
		"NAME",
		"PERF EXTENTS",
		"CAP TIER",
		"COPY",
		"MOVE",
		"MOVE DAYS",
		"ARCHIVE TIER",
		"ARCHIVE DAYS",
		"OFFLOAD ENCRYPT",
	)

	for _, i := range sobr.Data {
		t.AddLine(
			i.Name,
			len(i.PerformanceTier.PerformanceExtents),
			i.CapacityTier.Enabled,
			i.CapacityTier.CopyPolicyEnabled,
			i.CapacityTier.MovePolicyEnabled,
			i.CapacityTier.OperationalRestorePeriodDays,
			i.ArchiveTier.IsEnabled,
			i.ArchiveTier.ArchivePeriodDays,
			i.CapacityTier.Encryption.IsEnabled,
		)
	}

	fmt.Println("")
	t.Print()
	fmt.Println("")

	if save {
		utils.SaveData(&sobr, "sobr")
	}

}

func getAll(profile models.Profile) {
	fmt.Println("JOBS")
	getVbrJobs(profile)
	fmt.Println("")
	fmt.Println("PROXIES")
	getProxies(profile)
	fmt.Println("")
	fmt.Println("REPOS")
	getRepos(profile)
	fmt.Println("")
	fmt.Println("SOBR")
	getSobr(profile)
}

func getSessions(profile models.Profile) {
	sess := vhttp.GetData[vbrmodels.Sessions]("sessions?limit=20", profile)

	t := tabby.New()

	t.AddHeader("NAME", "TYPE", "CREATION TIME", "STATE", "PROGRESS", "RESULT")

	for _, s := range sess.Data {
		t.AddLine(
			s.Name,
			s.SessionType,
			s.CreationTime,
			s.State,
			s.ProgressPercent,
			s.Result.Result,
		)
	}

	fmt.Println("")
	t.Print()
	fmt.Println("")

	if save {
		utils.SaveData(&sess, "sessions")
	}

}

func getBackupObject(profile models.Profile) {
	objects := vhttp.GetData[vbrmodels.BackupObjects]("backupObjects", profile)

	t := tabby.New()

	t.AddHeader("NAME", "PLATFORM", "TYPE", "RESTORE POINTS")

	for _, o := range objects.Data {
		t.AddLine(
			o.Name,
			o.PlatformName,
			o.Type,
			o.RestorePointsCount,
		)
	}

	fmt.Println("")
	t.Print()
	fmt.Println("")

	if save {
		utils.SaveData(&objects, "backupObjects")
	}

}

func getInventory(profile models.Profile) {
	hosts := vhttp.GetData[vbrmodels.Hosts]("inventory/vmware/hosts", profile)

	println("Select Host to Inventory")
	for i, h := range hosts.Data {
		fmt.Println(i, h.InventoryObject.Name)
	}
	print("Selection: ")
	var selection string
	fmt.Scanln(&selection)

	i, err := strconv.Atoi(selection)
	utils.IsErr(err)

	sn := hosts.Data[i].InventoryObject.HostName

	ns := fmt.Sprintf("inventory/vmware/hosts/%v", sn)

	inventory := vhttp.GetData[vbrmodels.Inventory](ns, profile)

	t := tabby.New()

	fmt.Printf("%v inventory\n", ns)
	t.AddHeader("NAME", "TYPE", "SIZE")

	hostQty := 0
	vmQty := 0

	for _, o := range inventory.Data {
		if o.InventoryObject.Type == "Host" {
			hostQty += 1
		}
		if o.InventoryObject.Type == "VirtualMachine" {
			vmQty += 1
		}

		t.AddLine(
			o.InventoryObject.Name,
			o.InventoryObject.Type,
			o.Size,
		)
	}

	t.Print()

	fmt.Println("")
	fmt.Printf("Hosts: %v\n", hostQty)
	fmt.Printf("VMS: %v\n", vmQty)

	if save {
		st := fmt.Sprintf("%v-inventory", sn)
		utils.SaveData(&inventory, st)
	}

}
