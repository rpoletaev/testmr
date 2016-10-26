package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/html"
)

const maxProcess = 5

func main() {
	urlch := make(chan string)
	urls := strings.Split("https://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org/project/\nhttps://blog.golang.org\nhttps://blog.golang.org/", "\n")
	// semaphore := make(chan struct{}, maxProcess)
	wordCountChannel := make(chan uint, len(urls))
	var wg sync.WaitGroup
	go func() {
		for _, str := range urls {
			urlch <- str
		}
	}()

	for url := range urlch {
		// semaphore <- struct{}{}
		wg.Add(3)
		go func(u string) {
			defer wg.Done()
			wordCountChannel <- processURL(u)
		}(url)
		wg.Wait()
		// <-semaphore
	}

	var total uint
	go func() {
		for wc := range wordCountChannel {
			total += wc
		}
	}()

	fmt.Printf("total: %d\n", total)
}

func processURL(url string) (wordsCount uint) {
	if url = strings.TrimSpace(url); url != "" {
		resp, err := http.Get(url)
		if err != nil {
			println(err)
			return 0
		}

		page, perr := html.Parse(resp.Body)
		resp.Body.Close()
		if perr != nil {
			println(perr)
			return 0
		}

		wordsCount = getNodeWordsCount(page)
		fmt.Printf("%s %d\n", url, wordsCount)
	}
	return wordsCount
}

func getNodeWordsCount(node *html.Node) (counter uint) {
	if node.Type == html.TextNode && strings.TrimSpace(node.Data) != "" {
		counter += getStringWordsCount("go", node.Data)
	}

	for n := node.FirstChild; n != nil; n = n.NextSibling {
		counter += getNodeWordsCount(n)
	}

	return counter
}

func getStringWordsCount(word, source string) (counter uint) {
	fnc := func(r rune) bool {
		return !unicode.IsLetter(r)
	}

	words := strings.FieldsFunc(source, fnc)
	for _, w := range words {
		if strings.EqualFold(word, w) {
			counter++
		}
	}

	return counter
}
