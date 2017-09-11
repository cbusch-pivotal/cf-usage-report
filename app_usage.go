package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/labstack/echo"
	"github.com/palantir/stacktrace"
	"github.com/parnurzeal/gorequest"
)

//AppUsage array of orgs usage
type AppUsage struct {
	Orgs []OrgAppUsage `json:"orgs"`
}

//OrgAppUsage Single org usage
type OrgAppUsage struct {
	OrganizationGUID string    `json:"organization_guid"`
	OrgName          string    `json:"organization_name"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	AppUsages        []struct {
		SpaceGUID             string `json:"space_guid"`
		SpaceName             string `json:"space_name"`
		AppName               string `json:"app_name"`
		AppGUID               string `json:"app_guid"`
		InstanceCount         int    `json:"instance_count"`
		MemoryInMbPerInstance int    `json:"memory_in_mb_per_instance"`
		DurationInSeconds     int    `json:"duration_in_seconds"`
	} `json:"app_usages"`
}

// AppUsageReport handles the app-usage call validating the date
//  and executing the report creation
func AppUsageReport(c echo.Context) error {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return stacktrace.Propagate(err, "couldn't convert year to number")
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		return stacktrace.Propagate(err, "couldn't convert month to number")
	}
	usageReport, err := GetAppUsageReport(cfClient, year, month)

	if err != nil {
		return stacktrace.Propagate(err, "Couldn't get usage report")
	}
	return c.JSON(http.StatusOK, usageReport)
}

// GetAppUsageReport pulls the entire report together
func GetAppUsageReport(client *cfclient.Client, year int, month int) (*AppUsage, error) {
	if !(month >= 1 && month <= 12) {
		return nil, stacktrace.NewError("Month must be between 1-12")
	}

	// get a list of orgs within the foundation
	orgs, err := client.ListOrgs()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting list of orgs using client: %v", client)
	}

	report := AppUsage{}
	token, err := client.GetToken()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting token using client: %v", client)
	}

	// loop through orgs and get app usage report for each
	for _, org := range orgs {
		orgUsage, err := GetAppUsageForOrg(token, org, year, month)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed getting app usage for org: "+org.Name)
		}
		orgUsage.OrgName = org.Name
		report.Orgs = append(report.Orgs, *orgUsage)
	}

	return &report, nil
}

// GetAppUsageForOrg queries apps manager app_usages API for the orgs app usage information
func GetAppUsageForOrg(token string, org cfclient.Org, year int, month int) (*OrgAppUsage, error) {
	usageAPI := os.Getenv("CF_USAGE_API")
	target := &OrgAppUsage{}
	request := gorequest.New()
	resp, _, err := request.Get(usageAPI+"/organizations/"+org.Guid+"/app_usages?"+GenTimeParams(year, month)).
		Set("Authorization", token).TLSClientConfig(&tls.Config{InsecureSkipVerify: cfSkipSsl}).
		EndStruct(&target)
	if err != nil {
		return nil, stacktrace.Propagate(err[0], "Failed to get app usage report %v", org)
	}

	if resp.StatusCode != 200 {
		return nil, stacktrace.NewError("Failed getting app usage report %v", resp)
	}
	return target, nil
}
