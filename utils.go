package main

import (
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/p5quared/decanter/Autolab"
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
