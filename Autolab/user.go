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
func GetAutolab(client *http.Client, endpoint string, res interface{}) error {
	resp, err := client.Get(endpoint)
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
func GetSubmissions(client *http.Client, host, course, assessment string) ([]SubmissionsResponse, error) {
	var submissions []SubmissionsResponse
	err := GetAutolab(client, UrlSubmissions(host, course, assessment), &submissions)
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

func GetUserInfo(httpClient *http.Client, host string) (UserResponse, error) {
	var user UserResponse
	err := GetAutolab(httpClient, UrlUser(host), &user)
	if err != nil {
		return UserResponse{}, err
	}
	return user, nil
}

func GetUserCourses(httpClient *http.Client, host string) ([]CoursesResponse, error) {
	var courses []CoursesResponse
	err := GetAutolab(httpClient, UrlCourses(host), &courses)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func GetUserAssessments(httpClient *http.Client, host, course string) ([]AssessmentsResponse, error) {
	var assessments []AssessmentsResponse
	err := GetAutolab(httpClient, UrlAssessments(host, course), &assessments)
	if err != nil {
		return nil, err
	}
	return assessments, nil
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
