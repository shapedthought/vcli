package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/shapedthought/veeamcli/cloudmodels"
	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/shapedthought/veeamcli/vhttp"
)

// copied from azure assessment

var tpl *template.Template

var fm = template.FuncMap{
	"df": dataFormat,
	"dr": durationTime,
}

func dataFormat(c string) string {
	d := strings.Split(c, ".")
	if len(d) > 1 {
		dd := strings.Split(d[0], "T")
		t := fmt.Sprintf("%s %s", dd[1], dd[0])
		return t
	} else {
		return c
	}
}

func durationTime(c string) string {
	d := strings.Split(c, ".")
	return d[0]
}

func AzureAssessment(profile models.Profile) {
	ab := vhttp.GetData[cloudmodels.AboutServer]("system/about", profile)
	si := vhttp.GetData[cloudmodels.ServerInfo]("system/serverInfo", profile)

	tTo := time.Now()
	// one day back
	tFrom := tTo.AddDate(0, 0, -1)
	tString := tTo.Format("2006-01-02")
	fString := tFrom.Format("2006-01-02")

	ss := fmt.Sprintf("jobSessions?Types=PolicyBackup&Types=PolicySnapshot&FromUtc=%s&ToUtc=%s", fString, tString)
	ses := vhttp.GetData[cloudmodels.SessionInfo](ss, profile)

	// get the session ID from each session
	var sesId []cloudmodels.SessionId
	for _, s := range ses.Results {
		se := cloudmodels.SessionId{
			SessionId:  s.ID,
			PolicyName: s.BackupJobInfo.PolicyName,
		}
		sesId = append(sesId, se)
	}

	var sessLog []cloudmodels.SessionLog
	for _, s := range sesId {
		sesl := fmt.Sprintf("jobSessions/%v/log", s.SessionId)
		sr := vhttp.GetData[cloudmodels.SessionLog](sesl, profile)
		sessLog = append(sessLog, sr)
	}

	output := cloudmodels.OutputData{
		Version:       ab.ServerVersion,
		WorkerVersion: ab.WorkerVersion,
		AzureRegion:   si.AzureRegion,
		ServerName:    si.ServerName,
		StartTime:     tString,
		EndTime:       fString,
		SessionInfo:   ses,
		SessionLog:    sessLog,
	}
	nf, err := os.Create("index.html")
	utils.IsErr(err)
	tpl = template.Must(template.New("").Funcs(fm).ParseFiles("azuretpl.gohtml"))
	err = tpl.ExecuteTemplate(nf, "tpl.gohtml", output)
	utils.IsErr(err)
}
