package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

type urlData struct {
	url string
	patternCount int
	error string
}

func (this urlData) print() {
	if this.error == "" {
		fmt.Printf("Count for %s: %d\n", this.url, this.patternCount)
	} else {
		fmt.Printf("%s: %s\n", this.url, this.error)
	}
}

func getUrl(url string, pattern string) urlData {
	data := urlData{
		url: url,
		patternCount: 0,
		error: "",
	}

	res, err := http.Get(url)
	if err != nil {
		data.error = "is not avaliable"
		return data
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	data.patternCount = strings.Count(string(body), pattern)

	return data
}


func printResult(results <-chan urlData) {
	var totalCount int
	for r := range results {
		r.print()
		totalCount += r.patternCount
	}
	fmt.Printf("Total count: %d\n", totalCount)
}


func readUrls(urls chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := scanner.Text()
		if url == "" {
			continue
		}
		urls <- url
	}
	close(urls)
}


func fetchResult(urls chan string, results chan urlData, maxGoroutines int, pattern string) {
	goroutines := make(chan int, maxGoroutines)
	var wg sync.WaitGroup

	for url := range urls {
		goroutines <- 1
		wg.Add(1)
		go func(url string) {
			results <- getUrl(url, pattern)
			<-goroutines
			wg.Done()
		}(url)
	}
	wg.Wait()
	close(goroutines)
	close(results)
}


func main() {
	const pattern = "Go"
	const maxGoroutines = 5
	results := make(chan urlData)
	urls := make(chan string)

	go readUrls(urls)
	go fetchResult(urls, results, maxGoroutines, pattern)

	printResult(results)
}
