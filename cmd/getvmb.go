package cmd

import (
	"math"

	"github.com/cheynewallace/tabby"
	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/shapedthought/veeamcli/vbmmodels"
	"github.com/shapedthought/veeamcli/vhttp"
)

func options[T any](jsonF bool, yamlF bool, data *T, fn string) {
	if jsonF && !yamlF {
		utils.PrintJson(&data)
	}

	if yamlF && !jsonF {
		utils.PrintYaml(&data)
	}

	if save {
		utils.SaveData(&data, fn)
	}
}

func getVbmJobs(profile models.Profile) {

	// get the job states which have the name and id
	jobs := vhttp.GetData[vbmmodels.VbmJobs]("jobs", profile)

	if !jsonF && !yamlF {
		t := tabby.New()
		t.AddHeader("NAME", "TYPE", "ENABLED", "PROXY NAME", "REPO NAME", "LAST BACKUP", "STATUS")
		for _, i := range jobs {
			url := i.Links.BackupRepository.Href[3:]
			rn := vhttp.GetData[vbmmodels.VbmRepo](url, profile)
			urlp := rn.Links.Proxy.Href[3:]
			pn := vhttp.GetData[vbmmodels.VbmProxy](urlp, profile)
			t.AddLine(i.Name, i.BackupType, i.IsEnabled, pn.HostName, rn.Name, i.LastRun, i.LastStatus)
		}
		t.Print()
	}

	options(jsonF, yamlF, &jobs, "vbmJobs")

	// if jsonF && !yamlF {
	// 	utils.PrintJson(&jobs)
	// }

	// if yamlF && !jsonF {
	// 	utils.PrintYaml(&jobs)
	// }

	// if save {
	// 	utils.SaveData(&jobs, "vbmJobs")
	// }

}

func getVbmProxies(profile models.Profile) {
	proxies := vhttp.GetData[vbmmodels.VbmProxies]("Proxies?extendedView=true", profile)

	if !jsonF && !yamlF {
		t := tabby.New()
		t.AddHeader("HOSTNAME", "TYPE", "THREADS", "STATUS")

		for _, p := range proxies {
			t.AddLine(p.HostName, p.Type, p.ThreadsNumber, p.Status)
		}
		t.Print()
	}

	options(jsonF, yamlF, &proxies, "vbmProxies")
}

func getVbmRepos(profile models.Profile) {
	repos := vhttp.GetData[vbmmodels.VbmRepos]("BackupRepositories", profile)

	if !jsonF && !yamlF {
		t := tabby.New()
		t.AddHeader("NAME", "CAP GB", "USED GB", "DAILY RETENTION", "LONG TERM?", "RETENTION TYPE")

		for _, r := range repos {
			cap := r.CapacityBytes / int64(math.Pow(1024, 3))
			free := r.FreeSpaceBytes / int64(math.Pow(1024, 3))
			used := cap - free
			t.AddLine(r.Name, cap, used, r.DailyRetentionPeriod, r.IsLongTerm, r.RetentionType)
		}
		t.Print()
	}

	options(jsonF, yamlF, &repos, "vbmRepos")
}

func getVbmOrgs(profile models.Profile) {
	orgs := vhttp.GetData[vbmmodels.VbmOrgs]("Organizations", profile)

	if !jsonF && !yamlF {
		t := tabby.New()
		t.AddHeader("NAME", "TYPE", "ENABLED?", "EXCHANGE?", "SHAREPOINT?", "TEAMS?", "REGION")

		for _, o := range orgs {
			t.AddLine(o.Name, o.Type, o.IsBackedup, o.IsExchangeOnline, o.IsSharePointOnline, o.IsTeamsOnline, o.Region)
		}
		t.Print()
	}

	options(jsonF, yamlF, &orgs, "vbmOrgs")
}

func getVbmLicense(profile models.Profile) {
	lic := vhttp.GetData[vbmmodels.VbmLicense]("License", profile)

	if !jsonF && !yamlF {
		t := tabby.New()
		t.AddHeader("TYPE", "TO", "EXP DATE", "STATUS", "TOTAL", "USED")

		t.AddLine(lic.Type, lic.LicensedTo, lic.ExpirationDate, lic.Status, lic.TotalNumber, lic.UsedNumber)
		t.Print()
	}

	options(jsonF, yamlF, &lic, "vbmLicense")
}

func getVbmSessions(profile models.Profile) {
	sessions := vhttp.GetData[vbmmodels.VbmSessions]("JobSessions", profile)

	if !jsonF && !yamlF {
		t := tabby.New()
		t.AddHeader("CREATION TIME", "PROCESSED OBJS", "BOTTLENECK", "PROGRESS", "STATUS")

		for _, s := range sessions.Results {
			p := 0
			if s.Progress == 1 {
				p = 100
			} else {
				p = s.Progress
			}
			t.AddLine(s.CreationTime, s.Statistics.ProcessedObjects, s.Statistics.Bottleneck, p, s.Status)
		}
		t.Print()
	}

	options(jsonF, yamlF, &sessions, "vbmSessions")
}
