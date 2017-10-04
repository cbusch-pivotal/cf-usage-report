package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

type ApptioOrgAppUsage struct {
	ApptioAppUsages []ApptioAppUsages `json:"app_usages"`
}

type ApptioAppUsages struct {
	OrganizationGUID      string    `json:"organization_guid"`
	OrgName               string    `json:"organization_name"`
	PeriodStart           time.Time `json:"period_start"`
	PeriodEnd             time.Time `json:"period_end"`
	SpaceGUID             string    `json:"space_guid"`
	SpaceName             string    `json:"space_name"`
	AppName               string    `json:"app_name"`
	AppGUID               string    `json:"app_guid"`
	InstanceCount         int       `json:"instance_count"`
	MemoryInMbPerInstance int       `json:"memory_in_mb_per_instance"`
	DurationInSeconds     int       `json:"duration_in_seconds"`
}

func TestJsonOut(t *testing.T) {
	var file1 = "./reports/app_usages.json"
	//var file2 = "./reports/app_usages0.json"

	var apptioOrg ApptioOrgAppUsage
	org := getJSONTestFile(file1, apptioOrg)

	fmt.Printf(toJSON(org))

}

func getJSONTestFile(fileName string, apptioOrg ApptioOrgAppUsage) ApptioOrgAppUsage {
	raw, err := ioutil.ReadFile(fileName)
	check(err)

	var org OrgAppUsage
	json.Unmarshal(raw, &org)

	apptioOrg.ApptioAppUsages = make([]ApptioAppUsages, 0)

	for _, app := range org.AppUsages {
		appusage := ApptioAppUsages{
			OrganizationGUID:      org.OrganizationGUID,
			PeriodStart:           org.PeriodStart,
			PeriodEnd:             org.PeriodEnd,
			SpaceGUID:             app.SpaceGUID,
			SpaceName:             app.SpaceName,
			AppName:               app.AppName,
			AppGUID:               app.AppGUID,
			InstanceCount:         app.InstanceCount,
			MemoryInMbPerInstance: app.MemoryInMbPerInstance,
			DurationInSeconds:     app.DurationInSeconds,
		}
		apptioOrg.ApptioAppUsages = append(apptioOrg.ApptioAppUsages, appusage)

	}

	return apptioOrg
}

func GetOutputForApptio(usageReport *AppUsage) (ApptioOrgAppUsage, error) {

	var apptioOrg ApptioOrgAppUsage

	apptioOrg.ApptioAppUsages = make([]ApptioAppUsages, len(usageReport.Orgs))

	for _, orgs := range usageReport.Orgs {
		for _, app := range orgs.AppUsages {
			appusage := ApptioAppUsages{
				OrganizationGUID:      orgs.OrganizationGUID,
				PeriodStart:           orgs.PeriodStart,
				PeriodEnd:             orgs.PeriodEnd,
				SpaceGUID:             app.SpaceGUID,
				SpaceName:             app.SpaceName,
				AppName:               app.AppName,
				AppGUID:               app.AppGUID,
				InstanceCount:         app.InstanceCount,
				MemoryInMbPerInstance: app.MemoryInMbPerInstance,
				DurationInSeconds:     app.DurationInSeconds,
			}
			apptioOrg.ApptioAppUsages = append(apptioOrg.ApptioAppUsages, appusage)
		}

	}

	apptioreport := ApptioOrgAppUsage{}

	return apptioreport, nil

}

func toJSON(p interface{}) string {
	bytes, err := json.Marshal(p)
	check(err)
	return string(bytes)
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
