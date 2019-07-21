package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	crawlerconfig "./Config"
	crawler "./CrawlWrapper"
	urlprovider "./UrlProviderService"
	fwwc "./WriteFiles"
)

func applicationCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("")
		log.Println("Closing application")
		os.Exit(0)
	}()
}

func main() {
	applicationCloseHandler()
	globalDataChannel := make(chan []byte)
	globalWriteToFileChannel := make(chan []byte)
	urlsFromFileChannel := make(chan string)
	var appConfig *crawlerconfig.Config
	appConfig, err := crawlerconfig.Load()
	if err != nil {
		fmt.Println("Creating a new config file")
		fmt.Print("Enter path to workers directory: ")
		workersDirectoryPath, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		workersDirectoryPath = strings.Trim(workersDirectoryPath, "\n")
		appConfig.CrawlersDirectory = workersDirectoryPath

		fmt.Print("Enter path to directory with URL files: ")
		urlDirectoryPath, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		urlDirectoryPath = strings.Trim(urlDirectoryPath, "\n")
		appConfig.URLFilesDirectory = urlDirectoryPath

		err = fmt.Errorf("Invalid number of records (Minimum is 5)")
		var recordsPerFile int
		for err != nil {
			fmt.Print("Enter number of records per file: ")
			rpf, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			recordsPerFile, err = strconv.Atoi(strings.Trim(rpf, "\n"))
			if err != nil || recordsPerFile < 5 {
				err = fmt.Errorf("Invalid number of records (Minimum is 5)")
				fmt.Println(err)
			}
		}
		appConfig.RecordsPerFile = recordsPerFile
		err = appConfig.Save()
		if err != nil {
			log.Println("Could not save the config file, the app will prompt for config next time.")
		} else {
			log.Println("Saved config file to the root directory")
		}
	} else {
		log.Println("App config read from config file")
	}

	fileWriter := fwwc.GetInstance()

	files, err := ioutil.ReadDir(appConfig.CrawlersDirectory)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Found crawlers: ")
	for idx, f := range files {
		fmt.Printf("\t[%d]: %s\n", idx, f.Name())
	}
	fmt.Print("\nSelect crawler number to use: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	line = strings.Trim(line, "\n")
	crawlerNumber, err := strconv.Atoi(line)

	if err != nil || crawlerNumber >= len(files) {
		log.Fatalln("Invalid id")
	}

	crawlerPath := path.Join(appConfig.CrawlersDirectory, files[crawlerNumber].Name())

	crawlingWorker, err := crawler.GetCrawler(crawlerPath)

	if err != nil {
		log.Fatalln(err)
	}

	urlprovider.Init(appConfig.URLFilesDirectory)
	go urlprovider.GetUrls(urlsFromFileChannel)

	go func(crawledDataPipe chan<- []byte, crawlingUrlChannel <-chan string) {
		for url := range crawlingUrlChannel {
			crawledData, err := crawlingWorker.Crawl(url)
			if err != nil {
				log.Printf("Failed while crawling %s: '%s'.\n", url, err.Error())
			} else {
				go func(dataPipe chan<- []byte, data []byte) {
					dataPipe <- crawledData
				}(crawledDataPipe, crawledData)
				log.Println("Successfully crawled", url)
			}
		}
	}(globalDataChannel, urlsFromFileChannel)

	go fileWriter.Write(globalWriteToFileChannel, strings.Split(files[crawlerNumber].Name(), ".")[0])

	valuesCounter := 0
	var intermediateBuffer = []byte("[")
	for values := range globalDataChannel {
		intermediateBuffer = append(intermediateBuffer, values...)
		intermediateBuffer = append(intermediateBuffer, []byte(",")...)
		valuesCounter++
		if valuesCounter == appConfig.RecordsPerFile {
			intermediateBuffer = intermediateBuffer[:len(intermediateBuffer)-1]
			intermediateBuffer = append(intermediateBuffer, []byte("]")...)
			go func(ch chan<- []byte, dataChunk []byte) {
				ch <- dataChunk
			}(globalWriteToFileChannel, intermediateBuffer)
			intermediateBuffer = []byte("[")
			valuesCounter = 0
		}
	}
}
