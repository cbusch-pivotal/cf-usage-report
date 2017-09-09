package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"strconv"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/labstack/echo"
	"github.com/palantir/stacktrace"
	"github.com/parnurzeal/gorequest"
)

// TaskUsage array of orgs usage
type TaskUsage struct {
	Orgs []OrgTaskUsage `json:"orgs"`
}

// OrgTaskUsage Single org usage
type OrgTaskUsage struct {
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
	} `json:"service_usages"`
}

// TaskUsageReport handles the app-usage call validating the date
//  and executing the report creation
func TaskUsageReport(c echo.Context) error {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return stacktrace.Propagate(err, "couldn't convert year to number")
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		return stacktrace.Propagate(err, "couldn't convert month to number")
	}

	usageReport, err := GetTaskUsageReport(cfClient, year, month)

	if err != nil {
		return stacktrace.Propagate(err, "Couldn't get service usage report")
	}
	return c.JSON(http.StatusOK, usageReport)
	//	return c.JSON(http.StatusOK, "not yet implemented")
}

// GetTaskUsageReport pulls the entire report together
func GetTaskUsageReport(client *cfclient.Client, year int, month int) (*TaskUsage, error) {
	//if month > 12 || month < 1 {
	if !(month >= 1 && month <= 12) {
		return nil, stacktrace.NewError("Month must be between 1-12")
	}

	orgs, err := client.ListOrgs()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting list of orgs using client: %v", client)
	}

	report := TaskUsage{}
	token, err := client.GetToken()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting token using client: %v", client)
	}

	// loop through orgs and get app usage report for each
	for _, org := range orgs {
		orgUsage, err := GetTaskUsageForOrg(token, org, year, month)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed getting service usage for org: "+org.Name)
		}
		orgUsage.OrgName = org.Name
		report.Orgs = append(report.Orgs, *orgUsage)
	}

	return &report, nil
}

// GetTaskUsageForOrg queries apps manager app_usages API for the orgs app usage information
func GetTaskUsageForOrg(token string, org cfclient.Org, year int, month int) (*OrgTaskUsage, error) {
	usageAPI := os.Getenv("CF_USAGE_API")
	//cfSkipSsl := os.Getenv("CF_SKIP_SSL_VALIDATION") == "true"
	target := &OrgTaskUsage{}
	request := gorequest.New()
	resp, _, err := request.Get(usageAPI+"/organizations/"+org.Guid+"/app_usages?"+GenTimeParams(year, month)).
		Set("Authorization", token).TLSClientConfig(&tls.Config{InsecureSkipVerify: cfSkipSsl}).
		EndStruct(&target)
	if err != nil {
		return nil, stacktrace.Propagate(err[0], "Failed to get service usage report %v", org)
	}

	if resp.StatusCode != 200 {
		return nil, stacktrace.NewError("Failed getting service usage report %v", resp)
	}
	return target, nil
}
