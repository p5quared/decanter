package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/supabase-community/supabase-go"
	"golang.org/x/oauth2"
)

var lg = log.New(os.Stdout, "DEBUG MWARE: ", log.LstdFlags)

// MiddlewareRoundTripper is a struct that implements the http.RoundTripper interface
type MiddlewareRoundTripper struct {
	Proxied      http.RoundTripper
	RoundTripper http.RoundTripper
	preReq       []func(http.Request)
	postReq      []func(http.Response, error)
	logger       *log.Logger
}

// Set client.Transport
// WARN: All middleware must be safe for concurrent use.
// Middleware takes value types to enforce this somewhat.
func (lrt MiddlewareRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	done := make(chan struct{})
	var preReqWg sync.WaitGroup

	for _, f := range lrt.preReq {
		go func(f func(http.Request)) {
			preReqWg.Add(1)
			defer preReqWg.Done()
			if req != nil {
				f(*req)
			}
		}(f)
	}

	resp, err = lrt.RoundTripper.RoundTrip(req)

	var postReqWg sync.WaitGroup
	for _, f := range lrt.postReq {
		go func(f func(http.Response, error)) {
			postReqWg.Add(1)
			defer postReqWg.Done()
			if resp != nil {
				f(*resp, err)
			}
		}(f)
	}

	go func() {
		defer close(done)
		preReqWg.Wait()
		postReqWg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		lg.Println("Timed out waiting for preReq to finish.")
	}
	return
}

func OnBeforeRequest(f func(http.Request)) func(*MiddlewareRoundTripper) {
	return func(lrt *MiddlewareRoundTripper) {
		lrt.preReq = append(lrt.preReq, f)
	}
}

func OnAfterRequest(f func(http.Response, error)) func(*MiddlewareRoundTripper) {
	return func(lrt *MiddlewareRoundTripper) {
		lrt.postReq = append(lrt.postReq, f)
	}
}

func displayRequest(lg *log.Logger) func(http.Request) {
	return func(req http.Request) {
		lg.Printf("REQUEST TO: %s\n", req.URL)
		lg.Printf("Method: %s\n", req.Method)
		lg.Printf("Headers: %s\n", req.Header)
		lg.Printf("Body: %s\n", req.Body)
	}
}

func displayRespError(lg *log.Logger) func(http.Response, error) {
	return func(resp http.Response, err error) {
		if resp.StatusCode != 200 {
			lg.Printf("ERROR IN RESPONSE: %s\n", resp.Status)
			lg.Printf("Destination: %s\n", resp.Request.URL)
			lg.Println(err)
			lg.Println("Response: ", resp)
		}
	}
}

func NewMiddleware(transport *oauth2.Transport, opts ...func(*MiddlewareRoundTripper)) *MiddlewareRoundTripper {
	m := &MiddlewareRoundTripper{
		RoundTripper: transport,
		logger:       log.New(io.Discard, "No logging by default.", log.LstdFlags),
	}
	for _, f := range opts {
		f(m)
	}
	return m
}

func LiveMiddleware(c *http.Client) *MiddlewareRoundTripper {
	return NewMiddleware(
		c.Transport.(*oauth2.Transport),
	)
}

func WithTelemetry(c *http.Client) *MiddlewareRoundTripper {
	if PROD == "TRUE" {
		return ProdMiddleware(c)
	}
	return DebugMiddleware(c)
}

// NOTE: Expects that c is the result of oauth2.NewClient
func ProdMiddleware(c *http.Client) *MiddlewareRoundTripper {
	lg = log.New(io.Discard, "PRINT LOGGING DISABLED", log.LstdFlags)
	sb := newSupabaseTelemetry(supabaseUrlLive, supabaseKeyLive)
	return NewMiddleware(
		c.Transport.(*oauth2.Transport),
		OnBeforeRequest(sb.trackRequest),
		OnAfterRequest(sb.trackError),
	)
}

// NOTE: Expects that c is the result of oauth2.NewClient
func DebugMiddleware(c *http.Client) *MiddlewareRoundTripper {
	lg.Println("DEBUG MIDDLEWARE ACTIVE")
	//	sb := newSupabaseTelemetry(supabaseUrlDebug, supabaseKeyDebug)
	sb := newSupabaseTelemetry(supabaseUrlDebug, supabaseKeyDebug)
	return NewMiddleware(
		c.Transport.(*oauth2.Transport),
		OnBeforeRequest(displayRequest(lg)),
		OnAfterRequest(displayRespError(lg)),
		OnBeforeRequest(displayPath(lg)),

		OnBeforeRequest(sb.trackRequest),
		OnAfterRequest(sb.trackError),
	)
}

const ( // These are anon keys, so we good
	supabaseUrlLive = "https://wgbxigfylkllztzhjykd.supabase.co"
	supabaseKeyLive = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6IndnYnhpZ2Z5bGtsbHp0emhqeWtkIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MDg4MjY5NDcsImV4cCI6MjAyNDQwMjk0N30.vx_E8RClyIOsfGwAxa6O-J2cvvlT9ajcCf72_kwmI64"

	supabaseUrlDebug = "http://localhost:54321"
	supabaseKeyDebug = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0"
)

// NOTE: All Insert operations must be done with the `minimal` option
// to avoid the need to view the row. (RLS)
type SupabaseTelemetry struct {
	client *supabase.Client
}

func newSupabaseTelemetry(url, key string) *SupabaseTelemetry {
	c, err := supabase.NewClient(url, key, nil)
	if err != nil {
		log.Println("Error setting up telemetry:", err)
	}

	return &SupabaseTelemetry{
		client: c,
	}
}

func (t *SupabaseTelemetry) trackUser(req http.Request) {
}

func (t *SupabaseTelemetry) trackError(resp http.Response, err error) {
	if err != nil || resp.StatusCode != 200 {
		lg.Println("Logging error to supabase.")
		_, _, err := t.client.From("errors").Insert(map[string]interface{}{
			"url":    resp.Request.URL.Path,
			"status": resp.Status,
			"code":   resp.StatusCode,
			"error":  err,
		}, false, "", "minimal", "").Execute()
		if err != nil {
			log.Println(err)
		}
	}
}

func (t *SupabaseTelemetry) trackRequest(req http.Request) {
	url := req.URL.Path
	size, _ := httputil.DumpRequestOut(&req, true)

	lg.Println("Logging request to supabase.")
	_, _, err := t.client.From("requests").Insert(map[string]interface{}{
		"url":  url,
		"size": len(size),
	}, false, "", "minimal", "").Execute()
	if err != nil {
		lg.Println("Error logging request to supabase.")
		lg.Println(err)
	} else {
		lg.Println("Successfully logged request to supabase.")
	}
}

func (t *SupabaseTelemetry) trackResponse(resp http.Response, err error) {
}

func (t *SupabaseTelemetry) trackCoursesAndAssessments(req http.Request) {
	path := req.URL.Path
	split := strings.Split(path, "/")

	for i, s := range split {
		lg.Printf("Split[%d]: %s\n", i, s)
	}

	var courseID string
	var assessmentID string
	if len(split) > 4 && split[3] == "courses" {
		lg.Println("Logging to Supabase: Course")
		courseID = split[4]
	}
	if len(split) > 6 && split[5] == "assessments" {
		lg.Println("Logging to Supabase: Assessment")
		assessmentID = split[6]
	}
	if courseID == "" && assessmentID == "" {
		return
	}

	lg.Printf("Logging to Supabase: Course: %s, Assessment: %s\n", courseID, assessmentID)

	_, _, err := t.client.From("courses_assessments").Insert(map[string]interface{}{
		"course_id":     courseID,
		"assessment_id": assessmentID,
	}, false, "", "minimal", "").Execute()
	if err != nil {
		log.Println(err)
	}
}

func displayPath(lg *log.Logger) func(http.Request) {
	return func(req http.Request) {
		lg.Printf("PATH: %s\n", req.URL)

		split := strings.Split(req.URL.Path, "/")
		out := strings.Join(split, " | ")

		lg.Println("SPLIT:", out)
	}
}

type PocketbaseTelemetry struct {
	host string
}

func (pb PocketbaseTelemetry) pbTrackRequest(req http.Request) {
	fmt.Println("Logging request to pocketbase.")
	url := req.URL.String()
	size, _ := httputil.DumpRequestOut(&req, true)

	body := []byte(`{
		"url": "` + url + `",
		"size": ` + fmt.Sprint(len(size)) + `
	}`)

	endpoint := fmt.Sprintf("%s/api/collections/requests/records", pb.host)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		lg.Println("Error logging request to pocketbase.")
		lg.Println(err)
	}

	defer resp.Body.Close()
	fmt.Println("Successfully logged request to pocketbase.")
	fmt.Println(resp)
}

func newPocketbaseClient(host string) PocketbaseTelemetry {
	return PocketbaseTelemetry{
		host: host,
	}
}

var (
	pbUrlDebug = "http://localhost:8090"
	pbUrlLive  = "https://pocketbase.supabase.co"
)
