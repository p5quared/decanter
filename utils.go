package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/p5quared/decanter/Autolab"
	"golang.org/x/oauth2"
)

func filter[T any](elements []T, p func(T) bool) []T {
	var f []T
	for _, item := range elements {
		if p(item) {
			f = append(f, item)
		}
	}
	return f
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

func testSpinner(t spinner.Type) {
	spinner.New().
		Title("Example spinner!").
		Type(t).
		Action(func() {
			time.Sleep(5 * time.Second)
		}).Run()
}

func tokenExists(fs Autolab.TokenStore) bool {
	_, err := fs.Load()
	if err != nil {
		return false
	}
	return true
}

func newAutolabHTTPClient(authClient Autolab.AutolabOAuthClient, fs Autolab.TokenStore) *http.Client {
	ts := Autolab.NewAutolabTokenSource(fs, authClient)

	oauth2Client := oauth2.NewClient(context.Background(), ts)
	oauth2Client.Transport = WithTelemetry(oauth2Client)

	return oauth2Client
}

func pollLatestSubmission(c *http.Client, host, course, assessment string) (Autolab.SubmissionsResponse, error) {
	const timeout = 2 * time.Minute
	const pollInterval = 5 * time.Second
	for {
		select {
		case <-time.After(timeout):
			return Autolab.SubmissionsResponse{}, fmt.Errorf("timed out")
		case <-time.Tick(pollInterval):
			submissions, err := Autolab.GetSubmissions(c, host, course, assessment)
			if err != nil {
				fmt.Println(errorMsg("Polling failure: " + err.Error()))
				continue
			}
			var latest Autolab.SubmissionsResponse
			for _, sub := range submissions {
				if sub.Version > latest.Version {
					latest = sub
				}
			}
			if len(latest.Scores) > 0 {
				return latest, nil
			}
		}
	}
}
