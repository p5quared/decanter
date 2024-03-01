package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/speedata/optionparser"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/p5quared/decanter/Autolab"
	"golang.org/x/oauth2"
)

// By default, we're running in production mode.
const PROD = "TRUE"

const ( // I was told it was okay to put these here... :p
	host                 = "https://autolab.cse.buffalo.edu"
	decanterClientID     = "D4MZfAzZ27U121M2vwnHMEN6Cz-RMrQIKMlVjpEuKh8"
	decanterClientSecret = "fGVpQqJ0SdLGp7hfyN9wCn6VvuzU9djJfRklRPRQGGk"
)

func newAutolabClient(authClient Autolab.AutolabOAuthClient, fs Autolab.TokenStore) *http.Client {
	ts, err := Autolab.NewAutolabTokenSource(fs, authClient)
	if err != nil {
		panic(err)
	}

	oauth2Client := oauth2.NewClient(context.Background(), ts)
	oauth2Client.Transport = WithTelemetry(oauth2Client)

	return oauth2Client
}

// Big TODO: Caching user datda (like courses and due dates)
func main() {
	if PROD != "TRUE" {
		exclaim := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render
		fmt.Println(exclaim("RUNNING IN DEBUG MODE"))
	}

	op := optionparser.NewOptionParser()
	var file string
	op.On("-f NAME", "--file NAME", "Specify file. -f main.go", &file)

	var assessment string
	op.On("-a NAME", "--assessment NAME", "Specify an assessment. -a pa1", &assessment)

	var course string
	op.On("-c NAME", "--course NAME", "Specify a course. -c cse468-s24", &course)

	// Wait... This doesn't actually do anything (yet) lol
	var wait bool
	op.On("-w", "--wait", "Waits for additional info (if applicable). ex: 'submit -w' will wait for and display results. (NOT CURRENTLY IMPLEMENTED)", &wait)

	op.Command("setup", "Setup (authorize) a new device (you should only need to do this once).")
	op.Command("submit", "Submit to an assessment.Available flags: --course, --assessment, --file, --wait")
	op.Command("list", "List data. Args: courses|assessments|submissions|me")

	err := op.Parse()
	if err != nil {
		fmt.Println("Error parsing command: ", err)
		return
	}
	ex := op.Extra
	if len(ex) == 0 {
		fmt.Println("No command specified.")
		return
	}

	fs := NewFileTokenStore("auth.json")
	c := newAutolabClient(
		Autolab.NewAuthClient(decanterClientID, decanterClientSecret, host),
		fs,
	)

	if ex[0] == "setup" {
		interactiveSetup(Autolab.NewAuthClient(decanterClientID, decanterClientSecret, host), fs)
		return
	}

	// check that we have a token
	if !tokenExists(fs) {
		fmt.Println("No token found. Please run 'decanter setup' to authorize this device.")
		return
	}

	switch ex[0] {
	case "setup":
		interactiveSetup(Autolab.NewAuthClient(decanterClientID, decanterClientSecret, host), fs)
	// TODO: Should be a multipart form;
	// 1. Select course
	// 2. Select assessment
	// 3. Select file
	// Where each steps are pre-populated with
	// the users possible options, so they don't
	// have to remember the course/assessment names.
	// TODO: Save last used course/assessment to a cache
	// to avoid user having to re-enter it.
	// Prompt like "Resubmit [file] to [assessment]?"
	case "submit":
		if course == "" || assessment == "" || file == "" {
			// TODO: Form theme.
			huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Course").
						Description("Which course to submit for?\n'decanter list courses' to see options").
						Value(&course),
					huh.NewInput().
						Title("Assessment").
						Description("Which assessment to submit for?\n'decanter list assessments' to see options").
						Value(&assessment),
					huh.NewFilePicker().
						Title("Select a file:").
						Description("This file will be uploaded and submit to Autolab.").
						Value(&file),
				),
			).
				WithTheme(decanterFormStyle()).
				WithShowHelp(true).Run()
		}
		if file == "" || assessment == "" || course == "" {
			return // User cancelled
		}
		fmt.Printf("Submitting %s to %s...\n", file, assessment)
		var err error
		spinner.New().
			Title(" Submitting file to Autolab...").
			Style(spinStyle).
			Action(func() {
				_, err = Autolab.SubmitFile(c, host, course, assessment, file)
			}).Run()
		if err != nil {
			fmt.Println("Error submitting file: ", err)
		} else {
			var emphasis = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render
			fmt.Printf("%s %s to %s!\n", finished("Successfully submit"), emphasis(file), emphasis(assessment))
		}
	case "list":
		if len(ex) < 2 {
			fmt.Println("Invalid Usage: Please specify a list group.\nex: decanter list courses|assessments|me")
			return
		}
		switch ex[1] {
		case "courses":
			var courses []Autolab.CoursesResponse
			spinner.New().
				Title("Fetching course data...").
				Style(spinStyle).
				Action(func() {
					courses, _ = Autolab.GetUserCourses(c, host)
					courses = filterCourses(courses, func(c Autolab.CoursesResponse) bool {
						return c.Semester == "s24"
					})
				}).Run()

			fmt.Println(finished("Fetched course data"))
			displayCourseList(courses)
		case "assessments", "ass":
			var courses []Autolab.CoursesResponse
			spinner.New().
				Style(spinStyle).
				Title("Fetching assessments...").
				Action(func() {
					courses, _ = Autolab.GetUserCourses(c, host)
					courses = filter(courses, func(c Autolab.CoursesResponse) bool {
						return c.Semester == "s24"
					})
				}).Run()
			// Technically this is a lie.
			fmt.Println(finished("Fetched assessments"))
			displayAssessmentList(c, courses)
		case "submissions", "subs":
			if course == "" || assessment == "" {
				fmt.Println("Error: To view submissions, please pass a course and assessment")
				return
			}
			var submissions []Autolab.SubmissionsResponse
			spinner.New().
				Style(spinStyle).
				Title("Fetching submissions...").
				Action(func() {
					submissions, err = Autolab.GetSubmissions(c, host, course, assessment)
				}).Run()
			if err != nil {
				errStr := fmt.Sprintf("Error fetching submissions... :(\nCheck your arguments:\nCourse: %s\nAssessment: %s", course, assessment)
				fmt.Println(errStr)
				return
			}
			doneStr := fmt.Sprintf("Fetched submisions for %s", course)
			fmt.Println(finished(doneStr))

			// Grab latest
			var submission Autolab.SubmissionsResponse
			for _, sub := range submissions {
				if sub.Version > submission.Version {
					submission = sub
				}
			}

			displaySubmission(submission)
		case "me":
			var user Autolab.UserResponse
			spinner.New().
				Style(spinStyle).
				Title("Fetching user data...").
				Action(func() {
					user, _ = Autolab.GetUserInfo(c, host)
				}).Run()
			fmt.Println(finished("Fetched user data"))
			displayUserInfo(user)
		default:
			fmt.Println("Invalid list group.\nOptions: courses|assessments")
		}
	default:
		fmt.Println("Command not recognized.")
	}
}
