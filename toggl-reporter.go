package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jason0x43/go-toggl"
)

type TagsData struct {
	duration int64
	tasks    map[string]int64
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

func printReport(
	date string,
	report map[string]map[string]*TagsData,
	showDurationForEach bool,
) {
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
					"--- %s — %s ---\n\n",
					tags,
					formatMillis(tagsData.duration),
				)
				for description, duration := range tagsData.tasks {
					if showDurationForEach {
						fmt.Printf(
							"* %s — %s\n",
							description,
							formatMillis(duration),
						)
					} else {
						fmt.Printf("* %s\n", description)
					}
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
) map[string]map[string]*TagsData {
	report := make(map[string]map[string]*TagsData)
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
		_, ok := report[project]
		if !ok {
			report[project] = make(map[string]*TagsData)
		}
		_, ok = report[project][joinedTags]
		if !ok {
			report[project][joinedTags] = &TagsData{
				duration: 0,
				tasks:    make(map[string]int64),
			}
		}
		_, ok = report[project][joinedTags].tasks[timeEntry.Description]
		if !ok {
			report[project][joinedTags].tasks[timeEntry.Description] =
				timeEntry.Duration
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
	printWorkspaces := flag.Bool(
		"printWorkspaces",
		false,
		"Print workspaces instead of report",
	)
	showDurationForEach := flag.Bool(
		"showDurationForEach",
		false,
		"Show duration for each task",
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
			fmt.Printf("* %d — %s", workspace.ID, workspace.Name)
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
	report := composeReport(timeEntries)
	printReport(processedDate, report, *showDurationForEach)
}
