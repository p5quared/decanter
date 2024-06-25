package Autolab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type Autolab struct {
	c    *http.Client
	host string
}

func NewAutolab(client *http.Client) Autolab {
	host := "https://autolab.cse.buffalo.edu"
	return Autolab{client, host}
}

// These functions are used for user interaction with the Autolab API.
func (a Autolab) GetAutolab(endpoint string, res interface{}) error {
	resp, err := a.c.Get(endpoint)
	if err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return err
	}
	return nil
}

// Get submissions for a single assessment
func (a Autolab) GetSubmissions(course, assessment string) ([]SubmissionsResponse, error) {
	var submissions []SubmissionsResponse
	err := a.GetAutolab(UrlSubmissions(a.host, course, assessment), &submissions)
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func (a Autolab) GetUserInfo() (UserResponse, error) {
	var user UserResponse
	err := a.GetAutolab(UrlUser(a.host), &user)
	if err != nil {
		return UserResponse{}, err
	}
	return user, nil
}

func (a Autolab) GetUserCourses() ([]CoursesResponse, error) {
	var courses []CoursesResponse
	err := a.GetAutolab(UrlCourses(a.host), &courses)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func (a Autolab) GetUserAssessments(course string) ([]AssessmentsResponse, error) {
	var assessments []AssessmentsResponse
	err := a.GetAutolab(UrlAssessments(a.host, course), &assessments)
	if err != nil {
		return nil, err
	}
	return assessments, nil
}

func (a Autolab) PollLatestSubmission(course, assessment string) (SubmissionsResponse, error) {
	const timeout = 2 * time.Minute
	const pollInterval = 5 * time.Second
	for {
		select {
		case <-time.After(timeout):
			return SubmissionsResponse{}, fmt.Errorf("timed out")
		case <-time.Tick(pollInterval):
			submissions, err := a.GetSubmissions(course, assessment)
			if err != nil {
				return SubmissionsResponse{}, err
			}
			if len(submissions) < 1 {
				return SubmissionsResponse{}, fmt.Errorf("no submissions found")
			}

			var latest SubmissionsResponse
			for _, sub := range submissions {
				if sub.Version > latest.Version {
					latest = sub
				}
			}
			return latest, nil
		}
	}
}

func (a Autolab) SubmitFile(course, assmnt, fName string) (SubmitResponse, error) {
	endpoint := UrlSubmit(a.host, course, assmnt)

	file, err := os.Open(fName)
	if err != nil {
		return SubmitResponse{}, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("submission[file]", fName)
	if err != nil {
		return SubmitResponse{}, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return SubmitResponse{}, err
	}
	err = writer.Close()
	if err != nil {
		return SubmitResponse{}, err
	}

	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return SubmitResponse{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := a.c.Do(req)
	if resp.StatusCode != http.StatusOK {
		var respError struct {
			Error string `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&respError)
		if err != nil {
			return SubmitResponse{}, err
		}
		return SubmitResponse{}, fmt.Errorf("unexpected status code: %d, error: %s", resp.StatusCode, respError.Error)
	}
	if err != nil {
		return SubmitResponse{}, err
	}

	var submitResp SubmitResponse
	err = json.NewDecoder(resp.Body).Decode(&submitResp)
	if err != nil {
		return SubmitResponse{}, err
	}
	return submitResp, nil
}
