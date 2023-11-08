package googlesearch

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"google.golang.org/api/customsearch/v1"
)

const (
	// MaxResultsPerPage is the default number of search results per page
	MaxResultsPerPage int64 = 10
	// MaxResults is the maximum number of search results
	MaxResults int64 = 100
)

// Min returns the smaller of x or y.
func Min(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}

// SearchInput defines the input of the search task
type SearchInput struct {
	// Query: The search query.
	Query string `json:"query"`

	// TopK: The number of search results to return.
	TopK *int `json:"top_k,omitempty"`

	// IncludeLinkText: Whether to include the scraped text of the search web page result.
	IncludeLinkText *bool `json:"include_link_text,omitempty"`

	// IncludeLinkHtml: Whether to include the scraped HTML of the search web page result.
	IncludeLinkHtml *bool `json:"include_link_html,omitempty"`
}

type Result struct {
	// Title: The title of the search result, in plain text.
	Title string `json:"title"`

	// Link: The full URL to which the search result is pointing, e.g.
	// http://www.example.com/foo/bar.
	Link string `json:"link"`

	// Snippet: The snippet of the search result, in plain text.
	Snippet string `json:"snippet"`

	// LinkText: The scraped text of the search web page result, in plain text.
	LinkText string `json:"link_text"`

	// LinkHtml: The full raw HTML of the search web page result.
	LinkHtml string `json:"link_html"`
}

// SearchOutput defines the output of the search task
type SearchOutput struct {
	// Results: The search results.
	Results []*Result `json:"results"`
}

// Scrape the search results if needed
func scrapeSearchResults(searchResults *customsearch.Search, includeLinkText, includeLinkHtml bool) ([]*Result, error) {
	results := []*Result{}
	for _, item := range searchResults.Items {
		linkText, linkHtml := "", ""
		if includeLinkText || includeLinkHtml {
			// Make an HTTP GET request to the web page
			response, err := http.Get(item.Link)
			if err != nil {
				log.Printf("Error making HTTP GET request to %s: %v", item.Link, err)
				continue
			}
			defer response.Body.Close()

			// Parse the HTML content
			doc, err := goquery.NewDocumentFromReader(response.Body)
			if err != nil {
				fmt.Printf("Error parsing %s: %v", item.Link, err)
			}

			if includeLinkText {
				linkText = scrapeWebpageText(doc)
			}

			if includeLinkHtml {
				linkHtml, err = scrapeWebpageHtml(doc)
				if err != nil {
					log.Printf("Error scraping HTML from %s: %v", item.Link, err)
				}
			}
		}

		results = append(results, &Result{
			Title:    item.Title,
			Link:     item.Link,
			Snippet:  item.Snippet,
			LinkText: linkText,
			LinkHtml: linkHtml,
		})
	}
	return results, nil
}

// Search the web using Google Custom Search API and scrape the results if needed
func search(service *customsearch.Service, cseID string, query string, topK int64, includeLinkText bool, includeLinkHtml bool) ([]*Result, error) {
	if topK <= 0 || topK > MaxResults {
		return nil, fmt.Errorf("top_k must be between 1 and %d", MaxResults)
	}

	// Make the search request
	results := []*Result{}

	for start := int64(1); start <= topK; start += int64(MaxResultsPerPage) {
		searchNum := Min(topK-start+1, MaxResultsPerPage)
		searchResults, err := service.Cse.List().Cx(cseID).Q(query).Start(start).Num(searchNum).Do()
		if err != nil {
			return nil, err
		}
		rs, err := scrapeSearchResults(searchResults, includeLinkText, includeLinkHtml)
		if err != nil {
			return nil, err
		}
		results = append(results, rs...)
	}

	return results, nil
}

// Scrape the HTML content of a webpage
func scrapeWebpageHtml(doc *goquery.Document) (string, error) {
	return doc.Selection.Html()
}

// Scrape the text content of a webpage
func scrapeWebpageText(doc *goquery.Document) string {
	// Extract title, headers and paragraphs (h1, h2, h3, etc.)
	content := ""
	doc.Find("title, h1, h2, h3, h4, h5, h6, p, a").Each(func(index int, element *goquery.Selection) {
		text := element.Text()
		// Remove extra whitespace and newlines from the paragraph text
		text = strings.TrimSpace(text)
		// Append the text to the content string
		content += text
		tagName := strings.ToLower(element.Get(0).Data) // Get the tag name
		if tagName == "a" {
			content += " "
		} else {
			content += "\n"
		}
	})
	return content
}
