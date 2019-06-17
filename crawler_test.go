package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

type fakeFetcher map[string]fakeResult
type fakeResult []string

var fetcher = fakeFetcher{
	"u0": fakeResult{"u1", "u2", "u3"},
	"u1": fakeResult{"u2", "u3"},
	"u2": fakeResult{"u1", "u3"},
	"u3": fakeResult{"u1", "u2"},
}

func (f fakeFetcher) Fetch(url string, client HttpClient) ([]string, error) {
	if res, ok := f[url]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}

// a fake list of fetched url cache
var fakeCollectedUrls = &urlMap{
	cache: map[string][]string{
		"https://golang.org/": {"https://golang.org/pkg/",
			"https://golang.org/fmt/",
			"https://golang.org/os/"},
		"https://golang.org/pkg/": {"https://golang.org/fmt/",
			"https://golang.org/os/"},
		"https://golang.org/fmt/": {"https://golang.org/pkg/",
			"https://golang.org/os/"},
		"https://golang.org/os/": {"https://golang.org/pkg/",
			"https://golang.org/fmt/"},
	},
}

func TestPrettyPrintBuffer(t *testing.T) {

	var actual string
	var expected string

	actual = fakeCollectedUrls.PrintSiteMap("https://golang.org/", 1, 1)
	expected = "https://golang.org/"
	if actual != expected {
		t.Error("Mismatch in 1st case. Expected: ", expected, " \nactual: ", actual)
	}

	actual = fakeCollectedUrls.PrintSiteMap("https://golang.org/", 2, 1)
	expected =
		`https://golang.org/
	|__https://golang.org/pkg/
	|__https://golang.org/fmt/
	|__https://golang.org/os/`
	if actual != expected {
		t.Error("Mismatch in 1st case. \nExpected:\n ", expected, " \nactual:\n ", actual)
	}

	actual = fakeCollectedUrls.PrintSiteMap("https://golang.org/", 3, 1)
	expected =
		`https://golang.org/
	|__https://golang.org/pkg/
		|__https://golang.org/fmt/
		|__https://golang.org/os/
	|__https://golang.org/fmt/
		|__https://golang.org/pkg/
		|__https://golang.org/os/
	|__https://golang.org/os/
		|__https://golang.org/pkg/
		|__https://golang.org/fmt/`
	if actual != expected {
		t.Error("Mismatch in 1st case. \nExpected:\n ", expected, " \nactual:\n ", actual)
	}
}

func TestMetrics(t *testing.T) {
	fakeUrlStateMap := &urlStateMap{
		stateMap: map[string]urlState{
			"u1": DONE,
			"u2": DONE,
			"u3": DONE,
			"u4": LOADING,
			"u5": ERROR,
			"u6": LOADING,
			"u7": LOADING,
			"u8": ERROR,
			"u9": DONE,
		},
	}
	loaded, inprogress, err := fakeUrlStateMap.metrics()
	if loaded != 4 || err != 2 || inprogress != 3 {
		t.Error("Expected 4 loaded, 2 errors & 3 inprogresses; got loaded: ",
			loaded, " err: ", err, " inprogress: ", inprogress)
	}
}

func BenchmarkCrawler(b *testing.B) {
	runtime.GOMAXPROCS(1)
	benchUrls := [5]string{"https://golang.org", "https://google.com", "https://monzo.com", "https://example.com", "https://wiki.com"}
	depth := 3
	numBenchUrls := 5
	for i := 0; i < numBenchUrls; i++ {
		fmt.Println("iteration: ", i)
		fetcher := UrlFetcher{maxRetries: 3, baseUrl: benchUrls[i]}
		fmt.Println(benchUrls[i])
		Crawl(benchUrls[i], depth, fetcher)
	}
}

func BenchmarkCrawlerIterations(b *testing.B) {
	benchUrl := "https://golang.com"
	depth := 3
	for i := 0; i < 10; i++ {
		t1 := time.Now()
		fetcher := UrlFetcher{maxRetries: 3, baseUrl: benchUrl}
		Crawl(benchUrl, depth, fetcher)
		t2 := time.Since(t1)
		fmt.Println(benchUrl, "iteration: ", i, t2)
		for k := range visitedUrls.cache {
			delete(visitedUrls.cache, k)
		}
		for k := range urlStateCache.stateMap {
			delete(urlStateCache.stateMap, k)
		}
	}
}
