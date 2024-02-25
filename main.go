package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/speedata/optionparser"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/p5quared/decanter/Autolab"
	"golang.org/x/oauth2"
)

// By default, we're running in production mode.
var PROD = "TRUE"

var ( // I was told it was okay to put these here... :p
	host                 = "https://autolab.cse.buffalo.edu"
	decanterClientID     = "D4MZfAzZ27U121M2vwnHMEN6Cz-RMrQIKMlVjpEuKh8"
	decanterClientSecret = "fGVpQqJ0SdLGp7hfyN9wCn6VvuzU9djJfRklRPRQGGk"
)

// Complete device flow and cache token to disk
// TODO: Should ask to abort if token already exists
func setup(authClient Autolab.AutolabOAuthClient, fs Autolab.TokenStore) {
	// 1. DeviceAuth
	// 2. DeviceAccessCode
	// 3. Exchange
	// 4. Save
	// NOTE: Error on anything is fatal for now.

	var (
		dResp oauth2.DeviceAuthResponse
		dCode string
		token oauth2.Token
		err   error
	)

	spinner.New().
		Title("Initiating device flow...").
		Style(spinStyle).
		Action(func() {
			dResp, err = authClient.DeviceAuth()
		}).Run()
	if err != nil {
		fmt.Println("Error starting device flow: ", err)
		return
	}
	fmt.Println(finished("Starting device flow..."))
	fmt.Printf("\n    Visit %s and enter the code %s to authenticate.\n\n", url(dResp.VerificationURI), emph(dResp.UserCode))

	spinner.New().
		Title(" Polling server to for access code...").
		Type(spinner.Dots).
		Style(spinStyle).
		Action(func() {
			dCode, err = authClient.DeviceAccessCode(dResp)
		}).Run()
	if err != nil || dCode == "" {
		fmt.Println("Error getting code: ", err)
		return
	}
	fmt.Println(finished("Got code"))

	spinner.New().
		Title("Exchanging code for token...").
		Style(spinStyle).
		Action(func() {
			token, err = authClient.ExchangeCodeForToken(dCode)
		}).Run()
	if err != nil {
		fmt.Println("Error exchanging code for token: ", err)
		return
	}
	fmt.Println(finished("Exchanged code for token"))

	spinner.New().
		Title("Saving credentials...").
		Style(spinStyle).
		Action(func() {
			err = fs.Save(token)
		}).Run()
	if err != nil {
		fmt.Println("Error saving token: ", err)
		return
	}
	fmt.Println(finished("Saved token"))

	finStyle := lipgloss.NewStyle().
		Foreground(colorPrimary).Bold(true).
		PaddingTop(1).PaddingLeft(2)
	fmt.Println(finStyle.Render("Decanter setup complete!"))
}

func newAutolabClient(authClient Autolab.AutolabOAuthClient, fs Autolab.TokenStore) *http.Client {
	ts, err := Autolab.NewAutolabTokenSource(fs, authClient)
	if err != nil {
		panic(err)
	}

	oauth2Client := oauth2.NewClient(context.Background(), ts)
	oauth2Client.Transport = WithTelemetry(oauth2Client)

	return oauth2Client
}

func filterAssmts(assmts []Autolab.AssessmentsResponse, filter func(Autolab.AssessmentsResponse) bool) []Autolab.AssessmentsResponse {
	var filtered []Autolab.AssessmentsResponse
	for _, assmt := range assmts {
		if filter(assmt) {
			filtered = append(filtered, assmt)
		}
	}
	return filtered
}

func filterCourses(c []Autolab.CoursesResponse, filter func(Autolab.CoursesResponse) bool) []Autolab.CoursesResponse {
	var filtered []Autolab.CoursesResponse
	for _, course := range c {
		if filter(course) {
			filtered = append(filtered, course)
		}
	}
	return filtered
}

func displayCourseList(courses []Autolab.CoursesResponse) {
	var headerStyle = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Bold(true).
		Foreground(colorPrimary).
		PaddingTop(1).
		PaddingLeft(0)
	fmt.Println(headerStyle.Render("Your Current Courses (s24):"))

	var courseStyle = lipgloss.NewStyle().
		Align(lipgloss.Left).
		PaddingLeft(1)

	// TODO: Push an align course.Name to right
	for _, course := range courses {
		s := fmt.Sprintf("%s    (%s)", course.DisplayName, course.Name)
		fmt.Println(courseStyle.Render(s))
	}

	var helpStyle = lipgloss.NewStyle().
		Foreground(colorTextSubtle).PaddingLeft(1)

	helpStr := "[Course Name]               (Course ID)"
	fmt.Println(helpStyle.Render(helpStr))
}

func displayAssessmentList(httpClient *http.Client, courses []Autolab.CoursesResponse) {
	var headerStyle = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Bold(true).
		Foreground(colorPrimary).
		PaddingTop(1).
		PaddingLeft(0)
	fmt.Println(headerStyle.Render("Your Assessments:"))

	var OddRowStyle = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(colorTextPrimary)

	var EvenRowStyle = lipgloss.NewStyle().
		Inherit(OddRowStyle).
		Foreground(colorTextSubtle)

	t := table.New().
		Headers("Course", "Assessment", "Assigned", "Due Date", "Closed").
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(r, c int) lipgloss.Style {
			switch {
			case r == 0:
				return lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
			case r%2 == 0:
				return EvenRowStyle
			default:
				return OddRowStyle
			}
		})

	now := time.Now()
	var assessmentsFetched int
	for _, course := range courses {
		assessments, _ := Autolab.GetUserAssessments(httpClient, host, course.Name)
		for _, ass := range assessments {
			assessmentsFetched++
			// TODO: Parse dates
			t.Row(course.Name, ass.Name, ass.Assigned[5:10], ass.Due[5:10], ass.Closed[5:10])
		}
	}
	elapsed := time.Since(now)

	fmt.Println(t.Render())

	var helpStyle = lipgloss.NewStyle().
		Faint(true).
		Foreground(colorTextSubtle)
	// Not actually true, as we fetched the courses earlier and got assignments now.
	s := fmt.Sprintf("Fetched and processed %d courses and %d assessments in %v", len(courses), assessmentsFetched, elapsed.Truncate(time.Millisecond))
	fmt.Println(helpStyle.Render(s))
}

func displayUserInfo(user Autolab.UserResponse) {
	var emph = lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(colorPrimary).
		Bold(true)

	var body = lipgloss.NewStyle().
		Foreground(colorTextPrimary)

	// UB Autolab API doesn't appear to return anything else for now.
	// Could just be because I'm a transfer and have less info though...
	s1 := fmt.Sprintf("%s: %s %s", emph.Render("Name"), body.Render(user.FirstName), body.Render(user.LastName))
	s2 := fmt.Sprintf("%s: %s", emph.Render("Email"), body.Render(user.Email))
	fmt.Println()
	fmt.Println(s1)
	fmt.Println(s2)
}

func displaySubmission(submission Autolab.SubmissionsResponse) {
	title := emph("Latest Submission")
	fmt.Println(title)
	fmt.Println(" Version: ", submission.Version)
	fmt.Println(" Submitted: ", submission.Submitted)
	fmt.Println(" Filename: ", submission.Filename)

	var OddRowStyle = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(colorTextPrimary)

	var EvenRowStyle = lipgloss.NewStyle().
		Inherit(OddRowStyle).
		Foreground(colorTextSubtle)

	t := table.New().
		Headers("Problem", "Score").
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(r, c int) lipgloss.Style {
			switch {
			case r == 0:
				return lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
			case r%2 == 0:
				return EvenRowStyle
			default:
				return OddRowStyle
			}
		})
	scores := submission.Scores
	var keys []int
	for k := range scores {
		n, _ := strconv.Atoi(k)
		keys = append(keys, n)
	}

	sort.Ints(keys)

	for _, prob := range keys {
		score := scores[strconv.Itoa(prob)]
		prob_str := strconv.Itoa(prob)
		t.Row(prob_str, strconv.FormatFloat(score, 'f', -1, 32))
	}
	fmt.Println(t.Render())
}

func testSpinner(t spinner.Type) {
	spinner.New().
		Title("Example spinner!").
		Type(t).
		Action(func() {
			time.Sleep(5 * time.Second)
		}).Run()
}

// Big TODO: Caching user datda (like courses and due dates)
func main() {
	if PROD != "TRUE" {
		exclaim := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render
		fmt.Println(exclaim("RUNNING IN DEBUG MODE"))
	}

	op := optionparser.NewOptionParser()
	var file string
	op.On("-f NAME", "--file NAME", "Specify file.", &file)

	var assessment string
	op.On("-a NAME", "--assessment NAME", "Specify an assessment.", &assessment)

	var course string
	op.On("-c NAME", "--course NAME", "Specify a course.", &course)

	var wait bool
	op.On("-w", "--wait", "Waits for additional info (if applicable). ex: 'submit -w' will wait for and display results.", &wait)

	op.Command("setup", "Setup (authorize) a new device (you should only need to do this once).")
	op.Command("submit", "Submit to an assessment.Available flags: --course, --assessment, --file, --wait")
	op.Command("list", "List data. Args: courses|assessments|submissions|me")

	err := op.Parse()
	if err != nil {
		fmt.Println("Error parsing command: ", err)
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
		setup(Autolab.NewAuthClient(decanterClientID, decanterClientSecret, host), fs)
		return
	}

	// check that we have a token
	_, err = fs.Load()
	if err != nil {
		fmt.Println("Error: No token found. Please run 'decanter setup' to authorize a new device.")
		return
	}

	switch ex[0] {
	case "setup":
		setup(Autolab.NewAuthClient(decanterClientID, decanterClientSecret, host), fs)
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
			var emphasis = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
			fmt.Println()
			fmt.Printf("%s %s to %s!\n", finished("Successfully submit"), emphasis.Render(file), emphasis.Render(assessment))
		}
	case "list":
		if len(ex) < 2 {
			fmt.Println("Invalid Usage: Please specify a list group.\nex: decanter list courses|assessments|me")
			return
		}
		switch ex[1] {
		case "courses":
			start := time.Now()

			var courses []Autolab.CoursesResponse
			var numCourses int
			spinner.New().
				Title("Fetching course data...").
				Style(spinStyle).
				Action(func() {
					courses, _ = Autolab.GetUserCourses(c, host)
					numCourses = len(courses)
					courses = filterCourses(courses, func(c Autolab.CoursesResponse) bool {
						return c.Semester == "s24"
					})
				}).Run()

			fmt.Println(finished("Fetched course data"))
			displayCourseList(courses)
			var helpStyle = lipgloss.NewStyle().
				Foreground(colorTextSubtle)
			elapsed := time.Since(start)
			// Not actually true, as we fetched the courses earlier and got assignments now.
			s := fmt.Sprintf("\nFetched and processed %d courses in %v", numCourses, elapsed.Truncate(time.Millisecond))
			fmt.Println(helpStyle.Render(s))
		case "assessments", "ass":
			var courses []Autolab.CoursesResponse
			spinner.New().
				Style(spinStyle).
				Title("Fetching assessments...").
				Action(func() {
					courses, _ = Autolab.GetUserCourses(c, host)
					courses = filterCourses(courses, func(c Autolab.CoursesResponse) bool {
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
