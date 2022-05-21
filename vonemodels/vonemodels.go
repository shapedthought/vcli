package vonemodels

import "time"

type TriggeredAlarms struct {
	Items []struct {
		AlarmAssignment struct {
			ObjectID   int    `json:"objectId"`
			ObjectName string `json:"objectName"`
			ObjectType string `json:"objectType"`
		} `json:"alarmAssignment"`
		AlarmTemplateID  int       `json:"alarmTemplateId"`
		ChildAlarmsCount int       `json:"childAlarmsCount"`
		Description      string    `json:"description"`
		Name             string    `json:"name"`
		RepeatCount      int       `json:"repeatCount"`
		Status           string    `json:"status"`
		TriggeredAlarmID int       `json:"triggeredAlarmId"`
		TriggeredTime    time.Time `json:"triggeredTime"`
	} `json:"items"`
	TotalCount int `json:"totalCount"`
}
