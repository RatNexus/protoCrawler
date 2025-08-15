package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"sync/atomic"
)

func resolveRelativeURLs() {}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) { return }

func logError(message, rawCurrentURL string, err error) {
	log.Printf(" - error  : \"%s\" :ERROR: %s%s\n", rawCurrentURL, message, err)
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	var currentDepth uint // == 0 because go i awesome
	var pageCount int32

	if cfg.lo.logToFile {
		cfg.lf = cfg.lo.getLogFile()
		defer cfg.lf.Close()
	}

	if cfg.lo.logToScreen && cfg.lo.logToFile {
		multiWriter := io.MultiWriter(os.Stdout, cfg.lf)
		log.SetOutput(multiWriter)
	} else if cfg.lo.logToScreen {
		log.SetOutput(os.Stdout)
	} else if cfg.lo.logToFile {
		log.SetOutput(cfg.lf)
	} else {
		log.SetOutput(io.Discard)
	}

	if cfg.lo.doLogging && cfg.lo.doStart {
		// make this message variable depending on other logging options
		log.Printf("starting crawl of: \"%s\" with max depth of %d\n",
			cfg.baseURL, cfg.maxDepth)
		log.Print("\n----------\n\n")
	}

	cfg.concurrencyControl <- struct{}{}
	cfg.wg.Add(1)
	go cfg.internalCrawlPage(rawCurrentURL, currentDepth, &pageCount)
	cfg.wg.Wait()

	if cfg.lo.doLogging && cfg.lo.doEnd {
		log.Print("\n----------\n\n")
		log.Println("end of crawl")
		log.Print("\n----------\n\n")
	}

	// Todo: this should be a separate function
	if cfg.lo.doLogging && cfg.lo.doSummary {
		log.Println("Contents of the pages map:")
		cfg.mu.Lock() // log.Print is slower than making a map copy
		pages := make(map[string]int)
		for url, n := range cfg.pages {
			pages[url] = n
		}
		cfg.mu.Unlock()

		// thus no lock and direct use of cfg.pages here
		for url, n := range pages {
			log.Printf("%s ||| N: %d\n", url, n)
		}
	}

}

// Todo: add resoleRelativeURLs here
func (cfg *config) internalCrawlPage(rawCurrentURL string, currentDepth uint, pageCount *int32) {
	defer cfg.wg.Done()
	defer func() { <-cfg.concurrencyControl }()
	atomic.AddInt32(pageCount, 1)

	currentDepth += 1

	// enforce same host for base and current URLs, Part 1
	parsedCurrentUrl, err := url.Parse(rawCurrentURL)
	if err != nil {
		if cfg.lo.doLogging && cfg.lo.doErrors {
			logError("current url parse error: ", rawCurrentURL, err)
		}
		return
	}

	// normalise current URL
	currentUrl, err := normalizeURL(rawCurrentURL)
	if err != nil {
		if cfg.lo.doLogging && cfg.lo.doErrors {
			logError("normalise current url error: ", rawCurrentURL, err)
		}
		return
	}

	// increment or expand pages map of current URL
	cfg.mu.Lock()
	_, exists := cfg.pages[currentUrl]
	cfg.pages[currentUrl]++
	cfg.mu.Unlock()

	atomInt := atomic.LoadInt32(pageCount)
	if atomInt > cfg.maxPages && atomInt <= 0 {
		if cfg.lo.doLogging && cfg.lo.doPageAbyss {
			log.Printf("   aby-s: \"%s\"\n", rawCurrentURL)
		}
		return
	}

	if cfg.maxDepth > 0 && currentDepth > cfg.maxDepth {
		if cfg.lo.doLogging && cfg.lo.doDepthAbyss {
			log.Printf("   abyss: \"%s\"\n", rawCurrentURL)
		}
		return
	}

	doDepthLog := cfg.lo.doLogging && cfg.lo.doDepth
	if exists {
		if doDepthLog {
			log.Printf(" - depth %d: \"%s\"\n", currentDepth, rawCurrentURL)
		}
		return
	} else {
		if doDepthLog {
			log.Printf(" - DEPTH %d: \"%s\"\n", currentDepth, rawCurrentURL)
		}
	}

	// enforce same host for base and current URLs, Part 2
	if cfg.baseURL.Host != parsedCurrentUrl.Host {
		if cfg.lo.doLogging && cfg.lo.doErrors {
			err := fmt.Errorf(
				"host difference: Base: \"%s\", Current: \"%s\"",
				cfg.baseURL.Host, parsedCurrentUrl.Host)
			logError("", rawCurrentURL, err)
		}
		return
	}

	// parse links of current webpage
	webpage, err := getHTML(rawCurrentURL)
	if err != nil {
		if cfg.lo.doLogging && cfg.lo.doErrors {
			logError("getHTML error: ", rawCurrentURL, err)
		}
		return
	}
	gotURLs, err := getURLsFromHTML(webpage, rawCurrentURL)
	if err != nil {
		if cfg.lo.doLogging && cfg.lo.doErrors {
			logError("getURLsFromHTML error: ", rawCurrentURL, err)
		}
		return
	}

	// go deeper
	func() { <-cfg.concurrencyControl }()
	for _, v := range gotURLs {
		cfg.concurrencyControl <- struct{}{}
		cfg.wg.Add(1)

		if cfg.lo.doLogging && cfg.lo.doWidth {
			log.Printf("   width  : \"%s\"", v)
		}
		go cfg.internalCrawlPage(v, currentDepth, pageCount)
	}
	cfg.concurrencyControl <- struct{}{}
}
