package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/palantir/stacktrace"
	"github.com/parnurzeal/gorequest"
)

var cfClient *cfclient.Client

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

func main() {
	if os.Getenv("BASIC_USERNAME") == "" &&
		os.Getenv("BASIC_PASSWORD") == "" &&
		os.Getenv("CF_API") == "" &&
		os.Getenv("CF_USERNAME") == "" &&
		os.Getenv("CF_PASSWORD") == "" {
		log.Fatalf("Must set environment variables BASIC_USERNAME, BASIC_PASSWORD, CF_API, CF_USERNAME, CF_PASSWORD")
		return
	}
	_, err := SetupCfClient()
	if err != nil {
		log.Fatalf("Error setting up client %v", err)
		return
	}
	e := echo.New()
	e.GET("/app-usage/:year/:month", AppUsageReport)
	userBasic := os.Getenv("BASIC_USERNAME")
	passwordBasic := os.Getenv("BASIC_PASSWORD")
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == userBasic && password == passwordBasic {
			return true, nil
		}
		return false, nil
	}))
	e.Logger.Fatal(e.Start(":8080"))
}

// SetupCfClient logs into PCF
func SetupCfClient() (*cfclient.Client, error) {
	cfAPI := os.Getenv("CF_API")
	cfUser := os.Getenv("CF_USERNAME")
	cfPassword := os.Getenv("CF_PASSWORD")
	cfSkipSsl := os.Getenv("CF_SKIP_SSL_VALIDATION") == "true"

	c := &cfclient.Config{
		ApiAddress:        cfAPI,
		Username:          cfUser,
		Password:          cfPassword,
		SkipSslValidation: cfSkipSsl,
	}
	client, err := cfclient.NewClient(c)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error creating cf client")
	}
	cfClient = client
	return client, nil
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
	if month > 12 || month < 1 {
		return nil, stacktrace.NewError("Month must be between 1-12")
	}

	orgs, err := client.ListOrgs()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting list of apps using client: %v", client)
	}

	report := AppUsage{}
	token, err := client.GetToken()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed getting token using client: %v", client)
	}
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
	cfSkipSsl := os.Getenv("CF_SKIP_SSL_VALIDATION") == "true"
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

// GenTimeParams generates the from and to dates for the app_usages call to apps manager
func GenTimeParams(year int, month int) string {
	formatString := "2006-01-02"
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return "start=" + firstDay.Format(formatString) + "&end=" + lastDay.Format(formatString)
}
