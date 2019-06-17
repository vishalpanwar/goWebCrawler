package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

/*
Region: Type definitions
*/

/// <summary>
/// Maintains the cache of visited URLs along with their states
/// States being:
/// 1. LOADING
/// 2. Loaded
/// 3. Error
/// Contains a mutex to make sure the cache thread safe
/// </summary>
type urlMap struct {
	cache map[string][]string
	sync.Mutex
}

///<summary>
/// State enum to define the current state of the url to be fetched
///</summary>
type urlState int

const (
	DONE    urlState = 0
	LOADING urlState = 1
	ERROR   urlState = 2
	NA      urlState = 3
)

/// <summary>
/// Maintains the cache of visited URLs along with their states
/// States being:
/// 1. LOADING
/// 2. Loaded
/// 3. Error
/// Contains a mutex to make sure the cache thread safe
/// </summary>
type urlStateMap struct {
	stateMap map[string]urlState
	sync.Mutex
}

/// maps the child urls found during the parsing of a url
var visitedUrls = &urlMap{cache: make(map[string][]string)}

/// maintains the state of url during and after fetching from the fetcher
var urlStateCache = &urlStateMap{stateMap: make(map[string]urlState)}

var wg sync.WaitGroup

func (urlCache *urlMap) add(url string, urls []string) {
	urlCache.Lock()
	urlCache.cache[url] = urls
	urlCache.Unlock()
}

func (stateCache *urlStateMap) getState(url string) (*urlState, bool) {
	stateCache.Lock()
	defer stateCache.Unlock()
	state, ok := stateCache.stateMap[url]
	if ok {
		return &state, true
	}
	return nil, false
}

func (stateCache *urlStateMap) set(url string, state urlState) {
	stateCache.Lock()
	defer stateCache.Unlock()
	stateCache.stateMap[url] = state
}

/// <summary>
/// Prints the stats for each url crawled by the crawler
/// </summary>
func (stateCache *urlStateMap) metrics() (int, int, int) {
	loaded, inprogress, er := 0, 0, 0
	for _, state := range stateCache.stateMap {
		if state == DONE {
			loaded++
		}
		if state == LOADING {
			inprogress++
		}
		if state == ERROR {
			er++
		}
	}
	return loaded, inprogress, er
}

/// <summary>
/// Prints the urlMap in pretty site map manner
/// It prints the urls in a tree like structure
/// Ex: <URL1, level1>
///			|__<URL2, level2>
///					|___<URL3, level3>
///					|___<URL4, level3>
///		    |__<URL5, level2>
///					|__<URL6, level3>
/// </summary>
func (urlCache *urlMap) PrintSiteMap(url string, maxDepth int, space int) string {
	if space >= maxDepth {
		return url
	}

	res := ""

	//fmt.Println(url, " --> ", space, "  :  ", maxDepth)

	if newUrls, ok := urlCache.cache[url]; ok {
		for _, curUrl := range newUrls {
			res += "\n" + strings.Repeat("\t", space) + "|__" + urlCache.PrintSiteMap(curUrl, maxDepth, space+1)
		}
	}
	//fmt.Println("url: ", url, "dprth: ", maxDepth, "\n", res)
	return url + res
}

func (urlCache *urlMap) PrintSiteMapInfo() {
	for key, val := range urlCache.cache {
		fmt.Println(key, " --> ", val)
	}
}

/// <summerary>
/// Calls the fetcher for base url and enqueues the found urls for crawling
/// </summary>
func Crawl(baseUrl string, depth int, fetcher UrlFetcher) {

	defer wg.Done()

	if depth == 0 {
		return
	}

	_, inCache := urlStateCache.getState(baseUrl)
	if inCache == true {
		return
	}
	urlStateCache.set(baseUrl, LOADING)

	httpClient := GetClient{}
	urls, _err := fetcher.Fetch(baseUrl, httpClient)

	if _err != nil {
		fmt.Println("Unable to fetch the URL from ", baseUrl)
		urlStateCache.set(baseUrl, ERROR)
		return
	}

	urlStateCache.set(baseUrl, DONE)
	visitedUrls.add(baseUrl, urls)

	for _, url := range urls {
		wg.Add(1)
		go Crawl(url, depth-1, fetcher)
	}

	return
}

func main() {
	var (
		url        = flag.String("url", "https://monzo.com/", "site to crawl")
		maxDepth   = flag.Int("depth", 5, "maximum depth to crawl upto")
		outputFile = flag.String("output", "out.txt", "file to write the sitemap to")
	)
	flag.Parse()

	fetcher := UrlFetcher{maxRetries: 3, baseUrl: *url}
	start := time.Now()
	wg.Add(1)
	Crawl(*url, *maxDepth, fetcher)
	wg.Wait()
	timeTaken := time.Since(start)

	fmt.Println("Crawler took: ", timeTaken)

	start = time.Now()
	DONE, inprogress, er := urlStateCache.metrics()
	timeTaken = time.Since(start)

	fmt.Println("\nMetric aggregation took: ", timeTaken, "\nStats: LOADED =", DONE, "LOADING =", inprogress, "ERROR =", er)

	//fmt.Println(visitedUrls.PrintSiteMap(*url, 3, 1))

	start = time.Now()
	s := visitedUrls.PrintSiteMap(*url, *maxDepth, 1)
	err := ioutil.WriteFile(*outputFile, []byte(s), 0644)
	timeTaken = time.Since(start)
	if err != nil {
		fmt.Println("\nCannot write to the output file ", outputFile)
	} else {
		fmt.Println("\nWriting the url sitemap tree structure took: ", timeTaken)
	}

}
