package main

import (
	"context"
	"net/http"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/p5quared/decanter/Autolab"
	"golang.org/x/oauth2"
)

// Kind of a messy struct that carries around all the
// various Autolab-related clients and data.
type Decanter struct {
	Autolab.Autolab
	auth Autolab.AutolabOAuthClient
	ts   Autolab.TokenStore
	host string
}

func NewDecanter() Decanter {
	fs := NewFileTokenStore("auth.json")
	ac := Autolab.NewAuthClient(decanterClientID, decanterClientSecret, host)

	ts := Autolab.NewTokenSource(fs, ac)
	oauth2Client := oauth2.NewClient(context.Background(), ts)
	autolabClient := Autolab.NewAutolab(oauth2Client)

	return Decanter{autolabClient, ac, fs, host}
}

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

func (d Decanter) tokenExists() bool {
	_, err := d.ts.Load()
	if err != nil {
		return false
	}
	return true
}

func newAutolabHTTPClient(authClient Autolab.AutolabOAuthClient, fs Autolab.TokenStore) *http.Client {
	ts := Autolab.NewTokenSource(fs, authClient)

	oauth2Client := oauth2.NewClient(context.Background(), ts)
	// oauth2Client.Transport = WithTelemetry(oauth2Client)

	return oauth2Client
}
