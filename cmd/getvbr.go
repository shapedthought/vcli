package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/shapedthought/veeamcli/vbrmodels"
	"github.com/shapedthought/veeamcli/vhttp"
)

func getVbrJobStates(profile models.Profile) {
	url := "jobs/states"
	if len(nameF) > 0 {
		url += fmt.Sprintf("?nameFilter=%v", nameF)
	}

	jobData := vhttp.GetData[vbrmodels.VbrJobStates](url, profile)

	if !jsonF && !yamlF {
		t := tabby.New()

		t.AddHeader("NAME", "TYPE", "STATUS", "DESCRIPTION", "REPOSITORY", "OBJECT COUNT")
		for _, i := range jobData.Data {
			t.AddLine(i.Name, i.Type, i.Status, i.Description, i.RepositoryName, strconv.Itoa(i.ObjectsCount))
		}

		fmt.Println("")
		t.Print()
		fmt.Println("")

	}
	if jsonF && !yamlF {
		utils.PrintJson(&jobData)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&jobData)
	}

	if save {
		utils.SaveData(&jobData, "vbrJobStates")
	}

}

func getVbrJob(profile models.Profile, id string, name string) {
	if len(id) == 0 && len(name) == 0 {
		log.Fatal("name or id needs to be passed")
	}

	jId := id

	if len(name) > 0 {
		jobs := vhttp.GetData[vbrmodels.VbrJobs]("jobs", profile)
		check := false
		for _, i := range jobs.Data {
			if i.Name == name {
				jId = i.ID
				check = true
			}
		}
		if !check {
			log.Fatal("Job Name could not be found")
		}
	}

	jn := fmt.Sprintf("jobs/%v", jId)
	j := vhttp.GetData[vbrmodels.VbrJob](jn, profile)
	repos := vhttp.GetData[vbrmodels.Repos]("backupInfrastructure/repositories", profile)

	if !jsonF && !yamlF {
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
		t.Print()
	}

	if jsonF && !yamlF {
		utils.PrintJson(&j)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&j)
	}

	if save {
		utils.SaveData(&j, "vbrJob")
	}
}

func getVbrJobs(profile models.Profile) {
	url := "jobs"
	if len(nameF) > 0 {
		url += fmt.Sprintf("?nameFilter=%v", nameF)
	}
	// get the job states which have the name and id
	jobs := vhttp.GetData[vbrmodels.VbrJobs](url, profile)

	repos := vhttp.GetData[vbrmodels.Repos]("backupInfrastructure/repositories", profile)

	if !jsonF && !yamlF {
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
	}

	if jsonF && !yamlF {
		utils.PrintJson(&jobs)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&jobs)
	}

	if save {
		utils.SaveData(&jobs, "vbrJobs")
	}

}

func getProxies(profile models.Profile) {
	url := "backupInfrastructure/proxies"
	if len(nameF) > 0 {
		url += fmt.Sprintf("?nameFilter=%v", nameF)
	}
	p := vhttp.GetData[vbrmodels.Proxies](url, profile)

	if !jsonF && !yamlF {
		t := tabby.New()

		t.AddHeader("NAME", "DESCRIPTION", "TYPE", "MAX TASKS")

		for _, i := range p.Data {
			t.AddLine(i.Name, i.Description, i.Type, i.Server.MaxTaskCount)
		}

		fmt.Println("")
		t.Print()
		fmt.Println("")
	}
	if jsonF && !yamlF {
		utils.PrintJson(&p)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&p)
	}

	if save {
		utils.SaveData(&p, "proxies")
	}

}

func getRepos(profile models.Profile) {
	filtString := ""
	if len(nameF) > 0 {
		filtString += fmt.Sprintf("?nameFilter=%v", nameF)
	}
	rst := vhttp.GetData[vbrmodels.RepoStates]("backupInfrastructure/repositories/states", profile)

	var rd []vbrmodels.Repo

	for _, r := range rst.Data {
		st := fmt.Sprintf("backupInfrastructure/repositories/%v%v", r.ID, filtString)
		d := vhttp.GetData[vbrmodels.Repo](st, profile)
		rd = append(rd, d)
	}

	if !jsonF && !yamlF {
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
	}

	if jsonF && !yamlF {
		utils.PrintJson(&rst)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&rst)
	}

	if save {
		utils.SaveData(&rst, "repoStates")
	}

}

func getSobr(profile models.Profile) {
	url := "backupInfrastructure/scaleOutRepositories"
	if len(nameF) > 0 {
		url += fmt.Sprintf("?nameFilter=%v", nameF)
	}
	sobr := vhttp.GetData[vbrmodels.Sobr](url, profile)

	if !jsonF && !yamlF {
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
	}
	if jsonF && !yamlF {
		utils.PrintJson(&sobr)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&sobr)
	}

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
	url := "sessions?limit=20"
	if len(nameF) > 0 && !strings.Contains(nameF, "Failed") && !strings.Contains(nameF, "Success") {
		url += fmt.Sprintf("&nameFilter=%v", nameF)
	}
	sess := vhttp.GetData[vbrmodels.Sessions](url, profile)

	if !jsonF && !yamlF {
		t := tabby.New()

		t.AddHeader("NAME", "TYPE", "CREATION TIME", "STATE", "PROGRESS", "RESULT")

		for _, s := range sess.Data {
			if strings.Contains(nameF, "Failed") && s.Result.Result != "Failed" {
				continue
			}
			if strings.Contains(nameF, "Success") && s.Result.Result != "Success" {
				continue
			}
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
	}

	if jsonF && !yamlF {
		utils.PrintJson(&sess)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&sess)
	}

	if save {
		utils.SaveData(&sess, "sessions")
	}

}

func getBackupObject(profile models.Profile) {
	url := "backupObjects"
	if len(nameF) > 0 {
		url += fmt.Sprintf("&nameFilter=%v", nameF)
	}
	objects := vhttp.GetData[vbrmodels.BackupObjects](url, profile)

	if !jsonF && !yamlF {
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
	}
	if jsonF && !yamlF {
		utils.PrintJson(&objects)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&objects)
	}

	if save {
		utils.SaveData(&objects, "backupObjects")
	}

}

func getInventory(profile models.Profile) {
	filtString := ""
	if len(nameF) > 0 {
		filtString += fmt.Sprintf("&nameFilter=%v", nameF)
	}
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

	ns := fmt.Sprintf("inventory/vmware/hosts/%v%v", sn, filtString)

	inventory := vhttp.GetData[vbrmodels.Inventory](ns, profile)

	if !jsonF && !yamlF {
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
	}
	if jsonF && !yamlF {
		utils.PrintJson(&inventory)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&inventory)
	}

	if save {
		st := fmt.Sprintf("%v-inventory", sn)
		utils.SaveData(&inventory, st)
	}

}
