package Autolab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// These functions are used for user interaction with the Autolab API.
// TODO: None of these should panic.

// Get submissions for a single assessment
// TODO: This grabs an arbitrary score for now
func GetSubmissions(client *http.Client, host, course, assessment string) ([]SubmissionsResponse, error) {
	endpoint := UrlSubmissions(host, course, assessment)

	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var subs []SubmissionsResponse
	err = json.NewDecoder(resp.Body).Decode(&subs)
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func GetUserInfo(httpClient *http.Client, host string) (UserResponse, error) {
	endpoint := UrlUser(host)
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return UserResponse{}, err
	}

	var user UserResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return UserResponse{}, err
	}

	return user, nil
}

func GetUserCourses(httpClient *http.Client, host string) ([]CoursesResponse, error) {
	endpoint := UrlCourses(host)
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	var courses []CoursesResponse
	err = json.NewDecoder(resp.Body).Decode(&courses)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func GetUserAssessments(httpClient *http.Client, host, course string) ([]AssessmentsResponse, error) {
	endpoint := UrlAssessments(host, course)

	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	var assmts []AssessmentsResponse
	err = json.NewDecoder(resp.Body).Decode(&assmts)
	if err != nil {
		return nil, err
	}
	return assmts, nil
}

func SubmitFile(httpClient *http.Client, host, course, assmnt, fName string) (SubmitResponse, error) {
	endpoint := UrlSubmit(host, course, assmnt)

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

	resp, err := httpClient.Do(req)
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
