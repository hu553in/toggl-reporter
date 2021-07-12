package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jason0x43/go-toggl"
)

// Tags entity is collection of tasks tagged with them
// along with their summary duration
type Tags struct {
	duration int64
	tasks    []string
}

func contains(slice interface{}, item interface{}) bool {
	reflectValue := reflect.ValueOf(slice)
	if reflectValue.Kind() != reflect.Slice {
		panic("Invalid data type.")
	}
	for i := 0; i < reflectValue.Len(); i++ {
		if reflectValue.Index(i).Interface() == item {
			return true
		}
	}
	return false
}

func formatMillis(millis int64) string {
	duration := (time.Duration(millis) * time.Millisecond).Round(time.Second)
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute
	duration -= minutes * time.Minute
	seconds := duration / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func getTodayDate() string {
	datetime := time.Now()
	return getDateStringFromDatetime(datetime)
}

func getYesterdayDate() string {
	datetime := time.Now().Add(time.Duration(-24) * time.Hour)
	return getDateStringFromDatetime(datetime)
}

func getDateStringFromDatetime(datetime time.Time) string {
	return fmt.Sprintf(
		"%d-%02d-%02d",
		datetime.Year(),
		datetime.Month(),
		datetime.Day(),
	)
}

func processDate(date string) (string, error) {
	dateLower := strings.ToLower(date)
	if dateLower == "today" {
		return getTodayDate(), nil
	}
	if dateLower == "yesterday" {
		return getYesterdayDate(), nil
	}
	parsedDate, err := time.Parse(time.RFC3339, dateLower+"T00:00:00Z")
	return getDateStringFromDatetime(parsedDate), err
}

func printReport(date string, report map[string]map[string]*Tags) {
	fmt.Printf(
		"=========================================================\n\n"+
			"Report for %s\n\n"+
			"=========================================================\n\n",
		date,
	)
	var total int64 = 0
	if len(report) == 0 {
		fmt.Print("There is no data to print.\n\n")
	} else {
		for project, projectData := range report {
			fmt.Printf("++++++++ %s ++++++++\n\n", project)
			for tags, tagsData := range projectData {
				total += tagsData.duration
				fmt.Printf(
					"--- %s - %s ---\n\n",
					tags,
					formatMillis(tagsData.duration),
				)
				for _, task := range tagsData.tasks {
					fmt.Printf("* %s\n", task)
				}
				fmt.Println()
			}
		}
	}
	fmt.Printf(
		"=========================================================\n\n"+
			"Total: %s\n\n"+
			"=========================================================\n",
		formatMillis(total),
	)
}

func composeReport(
	timeEntries []toggl.DetailedTimeEntry,
	doNotMergeEqual bool,
) map[string]map[string]*Tags {
	report := make(map[string]map[string]*Tags)
	sort.Slice(timeEntries, func(i, j int) bool {
		return timeEntries[i].Start.Before(*timeEntries[j].Start)
	})
	for _, timeEntry := range timeEntries {
		project := timeEntry.Project
		if project == "" {
			project = "No project"
		}
		tags := timeEntry.Tags
		sort.Strings(tags)
		joinedTags := strings.Join(tags, ", ")
		if joinedTags == "" {
			joinedTags = "No tags"
		}
		{
			_, ok := report[project]
			if !ok {
				report[project] = make(map[string]*Tags)
			}
		}
		{
			_, ok := report[project][joinedTags]
			if !ok {
				report[project][joinedTags] = &Tags{
					duration: 0,
					tasks:    []string{},
				}
			}
		}
		if !contains(report[project][joinedTags].tasks, timeEntry.Description) ||
			doNotMergeEqual {
			report[project][joinedTags].tasks = append(
				report[project][joinedTags].tasks,
				timeEntry.Description,
			)
		}
		report[project][joinedTags].duration += timeEntry.Duration
	}
	return report
}

func main() {
	token := flag.String(
		"token",
		"",
		"Toggl API token (you can get it from your profile page)",
	)
	workspaceID := flag.String(
		"workspaceId",
		"",
		"Workspace ID (you can get it by running this app "+
			"with \"-printWorkspaces\" flag or just with \"-token\")",
	)
	date := flag.String(
		"date",
		"today",
		"Report date (can be: \"today\", \"yesterday\", \"YYYY-MM-DD\")",
	)
	doNotMergeEqual := flag.Bool(
		"doNotMergeEqual",
		false,
		"Do not merge tasks with equal descriptions",
	)
	printWorkspaces := flag.Bool(
		"printWorkspaces",
		false,
		"Print workspaces instead of report",
	)
	flag.Parse()
	processedDate, err := processDate(*date)
	if err != nil {
		fmt.Println("Date is invalid. Re-run app with \"-h\" or \"--help\" flag.")
		os.Exit(1)
	}
	if *token == "" {
		fmt.Println("Token is missing. Re-run app with \"-h\" or \"--help\" flag.")
		os.Exit(1)
	}
	toggl.DisableLog()
	session := toggl.OpenSession(*token)
	if *workspaceID == "" || *printWorkspaces {
		account, err := session.GetAccount()
		if err != nil {
			fmt.Println("Unable to get account.")
			os.Exit(1)
		}
		if !*printWorkspaces {
			fmt.Print("You did not entered workspace ID.\n\n")
		}
		fmt.Print("Your workspaces:\n\n")
		for _, workspace := range account.Data.Workspaces {
			fmt.Printf("* %d - %s", workspace.ID, workspace.Name)
		}
		fmt.Println()
		if !*printWorkspaces {
			os.Exit(1)
		}
		os.Exit(0)
	}
	parsedWorkspaceID, err := strconv.Atoi(*workspaceID)
	if err != nil {
		fmt.Println("Workspace ID is invalid.")
		os.Exit(1)
	}
	page := 1
	rawReport, err := session.GetDetailedReport(
		parsedWorkspaceID,
		processedDate,
		processedDate,
		page,
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	timeEntries := rawReport.Data
	if rawReport.TotalCount > rawReport.PerPage {
		pageCount := int(math.Ceil(
			float64(rawReport.TotalCount) / float64(rawReport.PerPage),
		))
		for i := 2; i <= pageCount; i++ {
			rawReportPage, err := session.GetDetailedReport(
				parsedWorkspaceID,
				processedDate,
				processedDate,
				i,
			)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			timeEntries = append(timeEntries, rawReportPage.Data...)
		}
	}
	report := composeReport(timeEntries, *doNotMergeEqual)
	printReport(processedDate, report)
}
