package cmd

import (
	"github.com/cheynewallace/tabby"
	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/vhttp"
	"github.com/shapedthought/veeamcli/vonemodels"
)

func triggeredAlarms(profile models.Profile) {
	alarms := vhttp.GetData[vonemodels.TriggeredAlarms]("alarms/triggered", profile)

	if !jsonF && !yamlF {
		t := tabby.New()
		t.AddHeader("NAME", "OBJ NAME", "REP COUNT", "STATUS", "TIME", "ID")
		for _, a := range alarms.Items {
			t.AddLine(a.Name, a.AlarmAssignment.ObjectName, a.RepeatCount, a.Status, a.TriggeredTime, a.TriggeredAlarmID)
		}
		t.Print()
	}

	options(jsonF, yamlF, &alarms, "voneTriggeredAlarms")
}
