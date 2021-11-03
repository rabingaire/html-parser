package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPageInfo(t *testing.T) {
	app := Setup()

	tests := []struct {
		Name     string
		Url      string
		Code     int
		Response GetPageInfoResponse
	}{
		{
			Name: "Get info of the webpage",
			Url:  "/api/v1/info?url=https://github.com/login",
			Code: http.StatusOK,
			Response: GetPageInfoResponse{
				HTMLVersion:            "5.0",
				PageTitle:              "Sign in to GitHub · GitHub",
				Headings:               map[string]int{"h1": 1},
				InternalLinksCount:     5,
				ExternalLinksCount:     4,
				InaccessibleLinksCount: 0,
				ContainsLoginForm:      true,
			},
		},
		{
			Name: "Get info of the html 4.01 webpage",
			Url:  "/api/v1/info?url=https://www.w3.org/TR/html401/intro/intro.html",
			Code: http.StatusOK,
			Response: GetPageInfoResponse{
				HTMLVersion:            "4.01",
				PageTitle:              "Introduction to HTML 4",
				Headings:               map[string]int{"h1": 1, "h2": 4, "h3": 14},
				InternalLinksCount:     80,
				ExternalLinksCount:     0,
				InaccessibleLinksCount: 19,
				ContainsLoginForm:      false,
			},
		},
		{
			Name: "Invalid URL",
			Url:  "/api/v1/info?url=https://github.com.login",
			Code: http.StatusInternalServerError,
		},
		{
			Name: "Invalid URL format (does not contain scheme)",
			Url:  "/api/v1/info?url=github.com",
			Code: http.StatusInternalServerError,
		},
		{
			Name: "URL not passed",
			Url:  "/api/v1/info",
			Code: http.StatusBadRequest,
		},
		{
			Name: "Empty URL passed",
			Url:  "/api/v1/info?url=",
			Code: http.StatusBadRequest,
		},
		{
			Name: "Invalid URL where server returns 404 status code",
			Url:  "/api/v1/info?url=https://github.com/jahdjsdfh",
			Code: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.Url, nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, tt.Code, rec.Code, "GET /info status code:\nwant  %+v\ngot  %+v", tt.Code, rec.Code)

			if rec.Code == http.StatusOK {
				var response GetPageInfoResponse
				json.Unmarshal(rec.Body.Bytes(), &response)

				assert.Equal(t, tt.Response, response, "GET /info response:\nwant  %+v\ngot  %+v", tt.Response, response)
			}
		})
	}
}

func TestParseHTML(t *testing.T) {
	u, err := url.Parse("https://example.com/")
	if err != nil {
		return
	}

	tests := []struct {
		Name     string
		RawHTML  string
		Response *GetPageInfoResponse
	}{
		{
			Name: "Parse basic html #1",
			RawHTML: `
			<html>
				<head>
					<title>World</title>
				</head>
				<body>
					<h1>h1</h1>
				</body>
			</html>
			`,
			Response: &GetPageInfoResponse{
				HTMLVersion:       "5.0",
				PageTitle:         "World",
				Headings:          map[string]int{"h1": 1},
				ContainsLoginForm: false,
				links:             make([]string, 0),
			},
		},
		{
			Name: "Parse basic html #2",
			RawHTML: `
			<html>
				<head>
					<title>Hello</title>
					<title>World</title>
				</head>
				<body>
					<h1>h1</h1>
					<h1>h1</h1>
					<h1>h1</h1>
					<h2>h2</h2>
					<div>
						<div>
							<h2>h2</h2>
						</div>
					</div>
				</body>
			</html>
			`,
			Response: &GetPageInfoResponse{
				HTMLVersion:       "5.0",
				PageTitle:         "Hello",
				Headings:          map[string]int{"h1": 3, "h2": 2},
				ContainsLoginForm: false,
				links:             make([]string, 0),
			},
		},
		{
			Name: "Parse basic html #3",
			RawHTML: `
			<html>
				<head>
					<title>Hello</title>
				</head>
				<body>
					<h1>h1</h1>
					<h1>h1</h1>
					<h1>h1</h1>
					<h2>h2</h2>
					<div>
						<div>
							<h2>h2</h2>
							<a href="/world">world</a>
						</div>
					</div>
				</body>
			</html>
			`,
			Response: &GetPageInfoResponse{
				HTMLVersion:            "5.0",
				PageTitle:              "Hello",
				Headings:               map[string]int{"h1": 3, "h2": 2},
				InternalLinksCount:     1,
				ExternalLinksCount:     0,
				InaccessibleLinksCount: 1,
				ContainsLoginForm:      false,
				links:                  []string{"https://example.com/world"},
			},
		},
		{
			Name: "Parse basic html #4",
			RawHTML: `
			<html>
				<head>
					<title>Hello</title>
				</head>
				<body>
					<h1>h1</h1>
					<h1>h1</h1>
					<h1>h1</h1>
					<h2>h2</h2>
					<div>
						<div>
							<h2>h2</h2>
							<a href="/world">world</a>
							<a href="https://facebook.com">facebook</a>
						</div>
					</div>
				</body>
			</html>
			`,
			Response: &GetPageInfoResponse{
				HTMLVersion:            "5.0",
				PageTitle:              "Hello",
				Headings:               map[string]int{"h1": 3, "h2": 2},
				InternalLinksCount:     1,
				ExternalLinksCount:     1,
				InaccessibleLinksCount: 1,
				ContainsLoginForm:      false,
				links:                  []string{"https://example.com/world", "https://facebook.com"},
			},
		},
		{
			Name: "Parse basic html #5",
			RawHTML: `
			<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
			<html>
				<head>
					<title>World</title>
				</head>
				<body>
					<h1>h1</h1>
				</body>
			</html>
			`,
			Response: &GetPageInfoResponse{
				HTMLVersion:       "4.01",
				PageTitle:         "World",
				Headings:          map[string]int{"h1": 1},
				ContainsLoginForm: false,
				links:             make([]string, 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			response, err := parseHTML(strings.NewReader(tt.RawHTML), u)
			if err == nil {
				assert.Equal(t, tt.Response, response, "parseHTML:\nwant  %+v\ngot  %+v", tt.Response, response)
			}
		})
	}
}
