package main

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/p5quared/decanter/Autolab"
	"golang.org/x/oauth2"
)

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

	var assessmentsFetched int
	for _, course := range courses {
		assessments, _ := Autolab.GetUserAssessments(httpClient, host, course.Name)
		for _, ass := range assessments {
			assessmentsFetched++

			// TODO: Align columns
			assTime := func(time_raw string) string {
				tt := Autolab.ParseTime(time_raw)
				return fmt.Sprintf("%s %d-%d", tt.Weekday().String()[:3], tt.Month(), tt.Day())
			}

			t.Row(course.Name, ass.Name, assTime(ass.Assigned), assTime(ass.Due), assTime(ass.Closed))
		}
	}
	fmt.Println(t.Render())
}

func displayTime(time_raw string) string {
	t := Autolab.ParseTime(time_raw)
	tSuffix := "AM"
	if t.Hour() >= 12 {
		tSuffix = "PM"
	}

	twelveHour := t.Hour()
	if twelveHour > 12 {
		twelveHour -= 12
	}

	return fmt.Sprintf("%s %d-%d (%d:%d%s)", t.Weekday(), t.Month(), t.Day(), twelveHour, t.Minute(), tSuffix)
}

func displaySubmission(submission Autolab.SubmissionsResponse) {

	title := emph("Latest Submission")
	fmt.Println(title)
	fmt.Println(" Version: ", submission.Version)
	fmt.Println(" Submitted: ", displayTime(submission.Submitted))
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
	keys := make([]string, 0, len(scores))
	for k := range scores {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		t.Row(k, fmt.Sprintf("%.2f", scores[k]))
	}
	fmt.Println(t.Render())
}

// Complete device flow and cache token to disk
// TODO: Should ask to abort if token already exists
func interactiveSetup(authClient Autolab.AutolabOAuthClient, fs Autolab.TokenStore) {
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
		Title(" Polling server for access code...").
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
	fmt.Println(finStyle.Render("Decanter setup complete! Try `decanter list me`"))
}

func areYouSure(msg, affirm, neg string) (ans bool) {
	huh.NewConfirm().
		Title(msg).
		Value(&ans).
		Affirmative(affirm).
		Negative(neg).
		WithTheme(decanterFormStyle()).
		Run()
	return
}
