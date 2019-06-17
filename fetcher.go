package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

/// Interface to return the http get response from http client
type HttpClient interface {
	Get(string) (*http.Response, error)
}

/// Returns silce of URLs from the parsed page
type Fetcher interface {
	Fetch(url string, client HttpClient) (urls []string, err error)
}

/// Implements HttpClient interface
type GetClient struct{}

/// Implements the fetcher interface and maintains the
/// maxRetry count for the fetcher instance
type UrlFetcher struct {
	maxRetries int
	baseUrl    string // needed for normalising relative paths & ignoring external paths
}

/// <summary>
/// returns a http response for the specified url
/// </summary>
func (h GetClient) Get(url string) (*http.Response, error) {
	httpClient := http.Client{}
	return httpClient.Get(url)
}

/// <summary>
/// creates a http client and fetches the response which is passed to a parser
/// </summary>
/// <parameter name="baseUrl">The parent url for which the page is parsed</parameter>
/// <parameter name="body">the page response body</parameter>
/// <return> retuens a set of parsed urls and error if any </return>
func (fetcher UrlFetcher) Fetch(url string, client HttpClient) ([]string, error) {

	var (
		response *http.Response
		err      error
	)

	retries := fetcher.maxRetries
	for retries >= 0 {
		response, err = client.Get(url)
		if err == nil {
			break
		}
		retries--
		// Sleep before retrying for 100ms
		time.Sleep(time.Millisecond * 100)
	}

	if err != nil && retries < 0 {
		return nil, err
	}

	defer response.Body.Close()

	urlSet, err := fetcher.htmlParser(url, response.Body)
	if err == nil {
		return urlSet, nil
	}

	return nil, err
}

/// <summary>
/// Parse the html page and get all the urls.
/// Filters the urls and add the unique ones to a string slice
/// </summary>
/// <parameter name="baseUrl">The parent url for which the page is parsed</parameter>
/// <parameter name="body">the page response body</parameter>
func (fetcher UrlFetcher) htmlParser(baseUrl string, body io.Reader) ([]string, error) {

	var urlCollection []string
	visUrlMap := make(map[string]bool)

	page := html.NewTokenizer(body)
	for {
		tokenType := page.Next()
		switch tokenType {
		case html.StartTagToken:
			token := page.Token()
			url, ok := fetcher.GetUrlFromHrefAttr(token)
			if ok == true {
				if url, err := fetcher.RefactorUrl(url, baseUrl); err == nil {
					if isUrlVisited(url, visUrlMap) == false {
						urlCollection = append(urlCollection, url)
						visUrlMap[url] = true
					} else {
						//fmt.Println("Repeated url: ", url)
					}
				}
			}
		case html.ErrorToken:
			if page.Err() == io.EOF {
				return urlCollection, nil
			}
		}
	}

	err := fmt.Errorf("No unique urls found for %s", baseUrl)
	return nil, err
}

/// <summary>
/// Get the url from the anchor tag
/// </summary>
func (fetcher UrlFetcher) GetUrlFromHrefAttr(token html.Token) (string, bool) {

	if token.DataAtom.String() == "a" {
		for _, attr := range token.Attr {
			if attr.Key == "href" {
				return attr.Val, true
			}
		}
	}
	return "", false
}

/// <summary>
/// Normalise all the relative urls and remove any external links
/// </summary>
/// <parameter name="link">new url parsed from the page and to be evaluated</parameter>
/// <parameter name="baseUrl">the parent url of the page parsed</parameter>
func (fetcher UrlFetcher) RefactorUrl(link string, baseUrl string) (string, error) {

	if len(link) == 0 ||
		link == baseUrl ||
		link == "/" ||
		strings.HasPrefix(link, "#") {
		return "", errors.New("invalid url type : " + link)
	}

	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	base, _ := url.Parse(fetcher.baseUrl)
	parsedLink := base.ResolveReference(parsedURL)
	if base.Host != parsedLink.Host {
		return "", errors.New("external urls are not crawled : " + link)
	}

	trimmedURL := strings.TrimSuffix(parsedLink.String(), "/")
	return trimmedURL, nil
}

/// <summary>
/// Check if a url is already visited in a non thread safe manner
/// as the map is local to the parse function
/// </summary>
/// <parameter name="url">url parsed from the page and to be evaluated</parameter>
/// <parameter name="visUrlMap">visited url map</parameter>
func isUrlVisited(url string, visUrlMap map[string]bool) bool {

	if _, isPresent := visUrlMap[url]; isPresent {
		return true
	}
	return false
}
