package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// setup logging output
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// main

	if len(os.Args) <= 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(os.Args) != 2 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}
	baseUrl := os.Args[1]
	maxDepth := 0 // dept 0 = inf
	log.Printf("starting crawl of: \"%s\" with max depth of %d\n", baseUrl, maxDepth)
	log.Print("\n----------\n\n")

	pages := make(map[string]int)
	crawlPage(baseUrl, baseUrl, pages, 0, maxDepth)

	log.Print("\n----------\n\n")
	log.Println("end of crawl")
	log.Print("\n----------\n\n")

	log.Println("Contents of the pages map:")
	for url, n := range pages {
		log.Printf("%s ||| N: %d\n", url, n)
	}
}
