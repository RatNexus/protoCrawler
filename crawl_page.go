package main

import (
	"fmt"
	"log"
	"net/url"
)

func resolveRelativeURLs() {}

// Todo: add resoleRelativeURLs here
func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int, currentDepth, maxDepth int) {
	currentDepth += 1
	if maxDepth > 0 && currentDepth > maxDepth {
		//log.Printf("   abyss: \"%s\"\n", rawCurrentURL)
		return
	}
	log.Printf(" - DEPTH %d\n", currentDepth)

	// enforce same host for both input URLs
	parsedBaseUrl, err := url.Parse(rawBaseURL)
	if err != nil {
		log.Println("Base url parse error: ", err)
		return
	}
	parsedCurrentUrl, err := url.Parse(rawCurrentURL)
	if err != nil {
		log.Println("Current url parse error: ", err)
		return
	}

	// normalise current URL
	currentUrl, err := normalizeURL(rawCurrentURL)
	if err != nil {
		log.Println("Normalise current url error: ", err)
		return
	}

	// increment or expand pages map of current URL
	_, exists := pages[currentUrl]
	pages[currentUrl]++
	if exists {
		return
	}

	if parsedBaseUrl.Host != parsedCurrentUrl.Host {
		log.Println(fmt.Printf(
			"Host difference: Base: \"%s\", Current: \"%s\"\n",
			parsedBaseUrl.Host, parsedCurrentUrl.Host))
		return
	}

	// parse links of current webpage
	webpage, err := getHTML(rawCurrentURL)
	if err != nil {
		log.Println("getHTML error: ", err)
	}
	gotURLs, err := getURLsFromHTML(webpage, rawCurrentURL)
	if err != nil {
		log.Println("getURLsFromHTML error: ", err)
	}

	// go deeper
	for _, v := range gotURLs {
		log.Printf("   width  : \"%s\"", v)
		crawlPage(rawBaseURL, v, pages, currentDepth, maxDepth)
	}
}
