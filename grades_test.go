package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	tests := []struct {
		Desc    string
		Handler func(http.ResponseWriter, *http.Request)
		Method  string
		Path    string
		Body    io.Reader
		Status  int
		Match   map[string]bool
	}{
		{
			Desc:    "blank submission",
			Handler: grades,
			Method:  "POST",
			Path:    "/grades",
			Body:    strings.NewReader(""),
			Status:  http.StatusOK,
			Match: map[string]bool{
				"error": true,
				"Fields missing from submission": true,
			},
		}, {
			Desc:    "incorrect submission: Affiliate is wrong",
			Handler: grades,
			Method:  "POST",
			Path:    "/grades",
			Body:    strings.NewReader(`{"Affiliate":"987","LastName":"smith"}`),
			Status:  http.StatusOK,
			Match: map[string]bool{
				"error": true,
				"No match for ID and last name": true,
			},
		}, {
			Desc:    "incorrect submission: LastName is wrong",
			Handler: grades,
			Method:  "POST",
			Path:    "/grades",
			Body:    strings.NewReader(`{"Affiliate":"123","LastName":"jones"}`),
			Status:  http.StatusOK,
			Match: map[string]bool{
				"error": true,
				"No match for ID and last name": true,
			},
		}, {
			Desc:    "incorrect submission: both Affiliate and LastName are wrong",
			Handler: grades,
			Method:  "POST",
			Path:    "/grades",
			Body:    strings.NewReader(`{"Affiliate":"456","LastName":"wilson"}`),
			Status:  http.StatusOK,
			Match: map[string]bool{
				"error": true,
				"No match for ID and last name": true,
			},
		}, {
			Desc:    "correct submission for: smith",
			Handler: grades,
			Method:  "POST",
			Path:    "/grades",
			Body:    strings.NewReader(`{"Affiliate":"123","LastName":"smith"}`),
			Status:  http.StatusOK,
			Match: map[string]bool{
				"FirstName":    true,
				"CurrentGrade": true,
			},
		}, {
			Desc:    "MARSHALLED correct submission for: smith",
			Handler: grades,
			Method:  "POST",
			Path:    "/grades",
			Body:    strings.NewReader(`{"Affiliate":"123","LastName":"smith"}`),
			Status:  http.StatusOK,
			Match: map[string]bool{
				"FirstName":    true,
				"CurrentGrade": true,
			},
		},
	}

	for _, test := range tests {
		record := httptest.NewRecorder()
		req, err := http.NewRequest(test.Method, test.Path, test.Body)
		if err != nil {
			t.Fatal(err)
		}

		test.Handler(record, req)

		if got, want := record.Code, test.Status; got != want {
			t.Errorf("%s: response code = %d, want %d", test.Desc, got, want)
		}
		for re, match := range test.Match {
			if got := regexp.MustCompile(re).Match(record.Body.Bytes()); got != match {
				t.Errorf("%s: %q ~ /%s/ = %v, want %v", test.Desc, record.Body, re, got, match)
			}
		}
		// t.Errorf("%s", record)
	}
}
