#### Usage

Concurrent web crawler which prints a site map of crawled website.

#### Dependencies
Download the following dependencies:
```
go get golang.org/x/net/html
```
#### Run 
To run the program, run the following command from inside the source repo
```
go build
./crawler -url https://monzo.com/ -depth 3 -output out.txt
```
The output sitemap is written to out.txt

#### Test
The project contains a bunch of test files for fetcher.go and crawler.go
Run ```go test``` command to run all the test cases

#### Benchmark
The design was benchmark against multiple approaches. ```Benchmark.txt``` contains the detailed info about the performance of various approaches.
```
go test -bench=.
```
#### Future Improvements
- Honor robots.txt 
- Use a global logger to create a log file and snapshot each event during the crawling.
- Get rid of global variables by creating a crawler class.
- Currently the crawler is super concurrent i.e. it creates a new go routine for each url in the parsed urls links, limit this by using a worker based approach where the number of active crawler goroutines can be ```X * no. of CPU cores```
- Persists the crawled links to disk in case of system crash and recreate the ```urlCacheMap``` from the logs.
- Add an expiry time to the crawled urls in case they need to be crawled again in future.
