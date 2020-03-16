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


func patternCounterForUrl(url string, commonChan chan struct{}, resChan chan urlDataType, wg *sync.WaitGroup) {
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
		<-commonChan
		resChan <- urlData
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	urlData.patternCount = strings.Count(string(body), pattern)

	<-commonChan
	resChan <- urlData
}


func main() {
	var wg sync.WaitGroup
	commonChan := make(chan struct{}, maxGoroutines)
	resChan := make(chan urlDataType)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := scanner.Text()
		if url == "" {
			continue
		}
		wg.Add(1)
		commonChan <- struct{}{}
		go patternCounterForUrl(url, commonChan, resChan, &wg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	totalCount := 0
	for res := range resChan {
		totalCount += res.patternCount
		fmt.Printf("Count for %s: %d\n", res.url, res.patternCount)
	}

	fmt.Println("The End")
	close(commonChan)
}
