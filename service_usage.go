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

// ServiceUsage array of orgs usage
type ServiceUsage struct {
	Orgs []OrgServiceUsage `json:"orgs"`
}

// OrgServiceUsage Single org usage
type OrgServiceUsage struct {
	OrganizationGUID string    `json:"organization_guid"`
	OrgName          string    `json:"organization_name"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	ServiceUsages    []struct {
		Deleted                 bool      `json:"deleted"`
		DurationInSeconds       float32   `json:"duration_in_seconds"`
		SpaceGUID               string    `json:"space_guid"`
		SpaceName               string    `json:"space_name"`
		ServiceInstanceGUID     string    `json:"service_instance_guid"`
		ServiceInstanceName     string    `json:"service_instance_name"`
		ServiceInstanceType     string    `json:"service_instance_type"`
		ServicePlanGUID         string    `json:"service_plan_guid"`
		ServicePlanName         string    `json:"service_plan_name"`
		ServiceName             string    `json:"service_name"`
		ServiceGUID             string    `json:"service_guid"`
		ServiceInstanceCreation time.Time `json:"service_instance_creation"`
		ServiceInstanceDeletion time.Time `json:"service_instance_deletion"`
	} `json:"service_usages"`
}

// ServiceUsageReport handles the app-usage call validating the date
//  and executing the report creation
func ServiceUsageReport(c echo.Context) error {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return stacktrace.Propagate(err, "couldn't convert year to number")
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		return stacktrace.Propagate(err, "couldn't convert month to number")
	}

	usageReport, err := GetServiceUsageReport(cfClient, year, month)

	if err != nil {
		return stacktrace.Propagate(err, "Couldn't get service usage report")
	}
	return c.JSON(http.StatusOK, usageReport)
}

// GetServiceUsageReport pulls the entire report together
func GetServiceUsageReport(client *cfclient.Client, year int, month int) (*ServiceUsage, error) {
	if !(month >= 1 && month <= 12) {
		return nil, stacktrace.NewError("Month must be between 1-12")
	}

	// get a list of orgs within the foundation
	orgs, err := client.ListOrgs()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting list of orgs using client: %v", client)
	}

	report := ServiceUsage{}
	token, err := client.GetToken()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting token using client: %v", client)
	}

	// loop through orgs and get app usage report for each
	for _, org := range orgs {
		orgUsage, err := GetServiceUsageForOrg(token, org, year, month)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed getting service usage for org: "+org.Name)
		}
		orgUsage.OrgName = org.Name
		report.Orgs = append(report.Orgs, *orgUsage)
	}

	return &report, nil
}

// GetServiceUsageForOrg queries apps manager app_usages API for the orgs app usage information
func GetServiceUsageForOrg(token string, org cfclient.Org, year int, month int) (*OrgServiceUsage, error) {
	usageAPI := os.Getenv("CF_USAGE_API")
	target := &OrgServiceUsage{}
	request := gorequest.New()
	resp, _, err := request.Get(usageAPI+"/organizations/"+org.Guid+"/service_usages?"+GenTimeParams(year, month)).
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
