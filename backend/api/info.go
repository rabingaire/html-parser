package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

type GetPageInfoResponse struct {
	HTMLVersion            string         `json:"html_version"`
	PageTitle              string         `json:"page_title"`
	Headings               map[string]int `json:"headings"`
	InternalLinksCount     int            `json:"internal_links_count"`
	ExternalLinksCount     int            `json:"external_links_count"`
	InaccessibleLinksCount int            `json:"inaccessible_links_count"`
	ContainsLoginForm      bool           `json:"contains_login_form"`

	links []string
}

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidURL          = errors.New("invalid URL")
)

func GetPageInfo(c *gin.Context) {
	reqURL := strings.Trim(c.Query("url"), "")
	if len(reqURL) <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidURL.Error()})
		return
	}

	u, err := url.Parse(reqURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidURL.Error()})
		return
	}

	res, err := http.Get(u.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServerError.Error()})
		return
	}

	response, err := parseHTML(res.Body, u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func parseHTML(r io.Reader, u *url.URL) (*GetPageInfoResponse, error) {
	response := &GetPageInfoResponse{
		Headings:          make(map[string]int),
		links:             make([]string, 0),
		ContainsLoginForm: false,
		HTMLVersion:       "5.0",
	}

	doc, err := html.Parse(r)
	if err != nil {
		return nil, ErrInternalServerError
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		// html version
		if n.Type == html.DoctypeNode && n.Data == "html" && len(n.Attr) > 0 {
			response.HTMLVersion = "4.01"
		}

		// get page title
		if len(response.PageTitle) == 0 &&
			n.Type == html.TextNode &&
			n.Parent.Type == html.ElementNode &&
			n.Parent.Data == "title" {
			response.PageTitle = n.Data
		}

		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6": // count all headings tags
				response.Headings[n.Data]++
			case "a": // get all the links
				for _, a := range n.Attr {
					if a.Key == "href" && len(a.Val) > 0 {
						// check if the link is internal
						if isInternalURL(a.Val) {
							response.links = append(
								response.links,
								mergePath(u, a.Val),
							)
							response.InternalLinksCount++
						} else {
							response.links = append(response.links, a.Val)
							response.ExternalLinksCount++
						}
						break
					}
				}
			case "input": // check if login form ins present
				for _, a := range n.Attr {
					if a.Key == "type" && a.Val == "password" {
						response.ContainsLoginForm = true
						break
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	message := make(chan int, len(response.links))
	for _, link := range response.links {
		go func(link string) {
			res, err := http.Get(link)
			if err != nil {
				message <- http.StatusInternalServerError
				return
			}
			defer res.Body.Close()
			message <- res.StatusCode
		}(link)
	}

	for i := 0; i < len(response.links); i++ {
		m := <-message
		if m != http.StatusOK {
			response.InaccessibleLinksCount++
		}
	}

	return response, nil
}

func isInternalURL(uri string) bool {
	u, err := url.Parse(uri)
	return !(err == nil && len(u.Scheme) > 0)
}

func mergePath(u *url.URL, link string) string {
	p := u.Path

	if len(path.Ext(p)) > 0 {
		p = path.Dir(p)
	}

	if path.IsAbs(link) {
		p = path.Clean(fmt.Sprintf("%s/%s", u.Host, link))
	} else {
		p = path.Join(p, link)
		p = path.Clean(fmt.Sprintf("%s/%s", u.Host, p))
	}

	return fmt.Sprintf("%s://%s", u.Scheme, p)
}
