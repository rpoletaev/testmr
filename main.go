package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/net/html"
)

const maxProcess = 5

func main() {
	fmt.Printf("%q\n", os.Args[1])
	urls := strings.Split(os.Args[1], "\n")
	// urls := strings.Split("https://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org\nhttps://golang.org/project/\nhttps://blog.golang.org\nhttps://blog.golang.org/", "\n")
	urlch := make(chan string, len(urls))
	for _, str := range urls {
		urlch <- str
	}
	close(urlch)

	wordCountChannel := make(chan uint, len(urls))
	var total uint

	for w := 0; w < maxProcess; w++ {
		go func() {
			for url := range urlch {
				wordCountChannel <- processURL(url)
			}
		}()
	}

	for i := 0; i < len(urls); i++ {
		total += <-wordCountChannel
	}

	fmt.Printf("total: %d\n", total)
}

type total struct {
	mu    sync.Mutex
	value uint
}

func (c *total) Total() uint {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func (c *total) Add(count uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = c.value + count
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
