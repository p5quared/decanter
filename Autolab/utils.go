package Autolab

import "fmt"

// TODO: Organize by scopes

// TODO: GET /writeup
// TODO: GET /handout
// TODO: GET /feedback
// TODO: POST /set_group_settings

// GET /user
type UserResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	School    string `json:"school"`
	Major     string `json:"major"`
	Year      string `json:"year"`
}

// GET /courses
type CoursesResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Semester    string `json:"semester"`
	LateSlack   int    `json:"late_slack"`
	GraceDays   int    `json:"grace_days"`
	AuthLevel   string `json:"auth_level"`
}

// GET /courses/:course_name/assessments
type AssessmentsResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Assigned    string `json:"start_at"`
	Due         string `json:"due_at"`
	Closed      string `json:"end_at"`
	Category    string `json:"category_name"`
}

type AssessmentDetailsResponse struct {
	Name                      string `json:"name"`
	DisplayName               string `json:"display_name"`
	Description               string `json:"description"`
	StartAt                   string `json:"start_at"`
	DueAt                     string `json:"due_at"`
	EndAt                     string `json:"end_at"`
	UpdateAt                  string `json:"updated_at"`
	MaxGraceDays              int    `json:"max_grace_days"`
	MaxSubmissions            int    `json:"max_submissions"`
	MaxUnpenalizedSubmissions int    `json:"max_unpenalized_submissions"`
	DsableHandins             bool   `json:"disable_handins"`
	CategoryName              string `json:"category_name"`
	GroupSize                 int    `json:"group_size"`
	WriteupFormat             string `json:"writeup_format"`
	HandoutFormat             string `json:"handout_format"`
	HasScoreboard             bool   `json:"has_scoreboard"`
	HasAutograder             bool   `json:"has_autograder"`
}

type SubmitResponse struct {
	Version  int    `json:"version"`
	Filename string `json:"filename"`
}

// Value is either float or 'unreleased' if score has not been
// released yet
type Scores map[string]float64
type SubmissionsResponse struct {
	Version   int    `json:"version"`
	Filename  string `json:"filename"`
	Submitted string `json:"created_at"`
	Scores    Scores `json:"scores"`
}

func UrlDFInit(host string) string {
	return host + "/oauth" + "/device_flow_init"
}

func UrlDFAuth(host string) string {
	return host + "/oauth" + "/device_flow_authorize"
}

func UrlToken(host string) string {
	return host + "/oauth" + "/token"
}

func UrlUser(host string) string {
	return host + "/api/v1" + "/user"
}

func UrlCourses(host string) string {
	return host + "/api/v1" + "/courses"
}

func UrlAssessments(host, course string) string {
	return host + "/api/v1" + "/courses" + "/" + course + "/assessments"
}

func UrlSubmissions(host, course, assessment string) string {
	return fmt.Sprintf("%s/api/v1/courses/%s/assessments/%s/submissions", host, course, assessment)
}

func UrlSubmit(host, course, assessment string) string {
	return fmt.Sprintf("%s/api/v1/courses/%s/assessments/%s/submit", host, course, assessment)
}
