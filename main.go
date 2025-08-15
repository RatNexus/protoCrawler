package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type loggingOptions struct {
	doLogging   bool
	doStart     bool
	doEnd       bool
	doSummary   bool
	doAbyss     bool
	doDepth     bool
	doWidth     bool
	doErrors    bool
	doPages     bool
	logsFolder  string
	doIdRoutine bool
	logName     string
	dateSuffix  string
	logToFile   bool
	logToScreen bool
}

func (lo *loggingOptions) setDefultLoggingOptions() {
	lo.doLogging = true
	lo.doStart = true
	lo.doEnd = true
	lo.doSummary = true
	lo.doAbyss = false
	lo.doWidth = true
	lo.doDepth = true
	lo.doErrors = true
	lo.doPages = true // not implemented

	lo.logToFile = false
	lo.logToScreen = true
	lo.logsFolder = "/tmp/protoCrawler/logs"
	lo.logName = "crawler"
	lo.dateSuffix = "2006-01-02_15:04:05"
	lo.doIdRoutine = true // not implemented
}
func (lo *loggingOptions) getLogFile() (lf *os.File) {
	var logFileName string
	if lo.dateSuffix != "" {
		currentTime := time.Now()
		tf := currentTime.Format(lo.dateSuffix)

		logFileName = fmt.Sprintf("%s_%s.log", lo.logName, tf)
	} else {
		logFileName = fmt.Sprintf("%s.log", lo.logName)
	}

	lo.logsFolder = strings.TrimRight(lo.logsFolder, "/")
	err := os.MkdirAll(lo.logsFolder, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	logPath := fmt.Sprintf("%s/%s", lo.logsFolder, logFileName)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return logFile
}

type config struct {
	pages              map[string]int
	mu                 *sync.Mutex // mutex for pages
	baseURL            *url.URL
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           uint
	maxDepth           uint
	lo                 *loggingOptions
	lf                 *os.File
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(os.Args) != 2 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	baseUrlStr := os.Args[1]
	cfg := config{}
	cfg.mu = &sync.Mutex{}
	cfg.pages = make(map[string]int)
	cfg.wg = &sync.WaitGroup{}
	var err error
	cfg.baseURL, err = url.Parse(baseUrlStr)
	if err != nil {
		log.Fatalf("Base url parse error: %v", err) // is this correct?
		os.Exit(1)
	}

	// setup logging
	cfg.lo = &loggingOptions{}
	cfg.lo.setDefultLoggingOptions()

	cfg.lo.doSummary = false
	cfg.lo.logToFile = false
	cfg.lo.logToScreen = true
	cfg.concurrencyControl = make(chan struct{}, 9)
	cfg.maxDepth = 3
	cfg.maxPages = 50

	// John Crawler
	cfg.crawlPage(baseUrlStr)
}
