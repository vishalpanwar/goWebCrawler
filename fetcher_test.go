package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"golang.org/x/net/html"
)

var mockSimpleFetcher = UrlFetcher{maxRetries: 1, baseUrl: "https://example.com"}

// we assume the standard http client works well and just test the fetcher
// properties - uniqueness of links and url clean-ups
type fakeHttpClient struct{}

func (f fakeHttpClient) Get(url string) (*http.Response, error) {

	// a fake http get response
	htmlBody := `<!DOCTYPE html>
		<head>
		  <title>Hello World</title>
		</head>
		<body class='section type-ccg'>
		<div class='hello_site'>
		    <a class='yolo' href='#main'>This is main cladd</a>
		    <header id='header' class='tolo1'>
		      <div class='header hello3'>
			<ul class='main-menu'>
		  <li>
		      <a href='/'>Home</a>
		    </li>
		  <li>
		      <a class='current' href='/blog/'>blog</a>
			</li>
			<li>
				<a class='current' aria-current='page' href='/blog/'>blog</a>
			</li>
		  <li>
		      <a href='/about/'>About</a>
		    </li>
			<li>
		      <a href='/blog'>About</a>
		    </li>
			<li>
		      <a href='/start-outreachy'>You should apply for Outreachy!</a>
			</li>
			<li>
		      <a href='/start-outreachy'>You should apply for Outreachy!</a>
			</li>
			<li>
		      <a href='/blog'>About</a>
		    </li>
			<li>
		      <a href='/start-outreachy'>You should apply for Outreachy!</a>
			</li>
			<li>
		      <a href='/start-outreachy'>You should apply for Outreachy!</a>
			</li>
			<li>
		      <a href='/Career'>Career</a>
		    </li>
			<li>
		      <a href='/start-outreachy'>Welcome to career</a>
			</li>
			<li>
		      <a href='/apply'>Welcome to career</a>
		    </li>
		  </ul>
			</header>
			<div class='hello_site'>
		    <a class='yolo' href='#main'>This is main cladd</a>
		    <header id='header' class='tolo1'>
			  <div class='header hello3'>
			  <div class='hello_site'>
		    <a class='yolo' href='/actions'>This is main cladd</a>
		    <header id='header' class='tolo1'>
		      <div class='header hello3'>
		/body>
		</html>`

	// NopCloser implments Reader and Closer, we can send out faked response
	// here
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(htmlBody)),
	}

	return resp, nil
}

func TestFetch(t *testing.T) {

	expectedUrls := []string{
		"https://example.com/blog",
		"https://example.com/about",
		"https://example.com/start-outreachy",
		"https://example.com/Career",
		"https://example.com/apply",
		"https://example.com/actions",
	}

	foundUrls, err := mockSimpleFetcher.Fetch("example.com", fakeHttpClient{})

	if err != nil {
		t.Error("Got error response from href tag")
	}

	if len(foundUrls) != len(expectedUrls) {
		t.Error("expected and found url size does not match.")
	}

	for i := 0; i < len(foundUrls); i++ {
		if foundUrls[i] != expectedUrls[i] {
			t.Error("expected and found url size not match. Expected: " + expectedUrls[i] + " Actual: " + foundUrls[i])
		}
	}
}

func TestGetAnchorHrefAttr(t *testing.T) {

	expectedUrl := "https://example.com/docs/"
	anchorToken := html.Token{
		Type:     html.StartTagToken,
		DataAtom: 0x1,
		Data:     "a",
		Attr:     []html.Attribute{html.Attribute{Key: "href", Val: expectedUrl}},
	}

	foundUrl, ok := mockSimpleFetcher.GetUrlFromHrefAttr(anchorToken)

	if ok != true {
		t.Error("Got fasle response from href tag")
	}
	if foundUrl != expectedUrl {
		t.Error("expected and found url does not match. Expected: " + expectedUrl + " Actual: " + foundUrl)
	}
}

func TestCleanUpUrl(t *testing.T) {

	parentLink := "https://example.com/doc"

	_, err := mockSimpleFetcher.RefactorUrl("", parentLink)
	if err == nil {
		t.Error("url cleanup error : empty urls should be removed")
	}

	_, err = mockSimpleFetcher.RefactorUrl("/", parentLink)
	if err == nil {
		t.Error("url cleanup error : links to homepage should be removed")
	}

	_, err = mockSimpleFetcher.RefactorUrl("#question", parentLink)
	if err == nil {
		t.Error("url cleanup error : self-loops should be removed")
	}

	_, err = mockSimpleFetcher.RefactorUrl(parentLink, parentLink)
	if err == nil {
		t.Error("url cleanup error : self-loops should be removed")
	}

	_, err = mockSimpleFetcher.RefactorUrl("https://twitter.com", parentLink)
	if err == nil {
		t.Error("url cleanup error : external urls should be removed")
	}

	url, err := mockSimpleFetcher.RefactorUrl("/help", parentLink)
	if err != nil {
		t.Error("error while retreiving the url")
	}

	if url != "https://example.com/help" {
		t.Error("wrong url")
	}
}
