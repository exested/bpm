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

type urlDataType struct {
	url string
	patternCount int
}

const pattern = "Go"
const maxGoroutines = 5


func patternCounterForUrl(
	url string,
	goroutines <-chan struct{},
	resultChan chan<- urlDataType,
	wg *sync.WaitGroup,
	) {
	urlData := urlDataType{
		url: url,
		patternCount: 0,
	}
	defer func() {
		wg.Done()
	}()

	res, err := http.Get(url)
	if err != nil {
		fmt.Println("URL",  url, "not available")
		<-goroutines
		resultChan <- urlData
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	urlData.patternCount = strings.Count(string(body), pattern)

	resultChan <- urlData
	<-goroutines
}

func printResult(resChan <-chan urlDataType, totalCountChan chan<- int)  {
	totalCount := 0
	for res := range resChan {
		totalCount += res.patternCount
		fmt.Printf("Count for %s: %d\n", res.url, res.patternCount)
	}
	totalCountChan <- totalCount
	close(totalCountChan)
}

func main() {
	var wg sync.WaitGroup
	goroutines := make(chan struct{}, maxGoroutines)
	resultChan := make(chan urlDataType)
	totalCountChan := make(chan int, 1)

	defer func() {
		wg.Wait()
		close(resultChan)
		fmt.Printf("Total count: %d\n", <-totalCountChan)
	}()

	go printResult(resultChan, totalCountChan)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := scanner.Text()
		if url == "" {
			continue
		}
		wg.Add(1)
		goroutines <- struct{}{}
		go patternCounterForUrl(url, goroutines, resultChan, &wg)
	}
	close(goroutines)

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
