package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	crawlerconfig "./Config"
	crawler "./CrawlWrapper"
	fwwc "./WriteFiles"
)

func main() {
	globalDataChannel := make(chan []byte)
	globalWriteToFileChannel := make(chan []byte)
	var appConfig *crawlerconfig.Config
	appConfig, err := crawlerconfig.Load()
	if err != nil {
		fmt.Println("Creating a new config file")
		fmt.Print("Enter path to workers directory: ")
		workersDirectoryPath, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		workersDirectoryPath = strings.Trim(workersDirectoryPath, "\n")
		appConfig.CrawlersDirectory = workersDirectoryPath

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

	var crawlingWorker crawler.Crawler
	go func(ch chan<- []byte, crawlingScriptPath string) {
		crawlingWorker = &crawler.Worker{Language: "python", ScriptPath: crawlingScriptPath}
		crawledData, err := crawlingWorker.Crawl("https://oyaop.com/opportunity/scholarships-and-fellowships/fully-funded-icwa-fellowship-program-2019-in-usa/")
		if err != nil {
			log.Fatal(err.Error())
		}
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		ch <- crawledData
		log.Println("Successfully crawled", "https://oyaop.com/opportunity/scholarships-and-fellowships/fully-funded-icwa-fellowship-program-2019-in-usa/")
	}(globalDataChannel, crawlerPath)

	go fileWriter.Write(globalWriteToFileChannel, strings.Split(files[crawlerNumber].Name(), ".")[0])

	valuesCounter := 0
	var intermediateBuffer = []byte("[")
	for values := range globalDataChannel {
		if valuesCounter == 5 {
			intermediateBuffer = intermediateBuffer[:len(intermediateBuffer)-1]
			intermediateBuffer = append(intermediateBuffer, []byte("]")...)
			go func(ch chan<- []byte, dataChunk []byte) {
				ch <- dataChunk
			}(globalWriteToFileChannel, intermediateBuffer)
			intermediateBuffer = []byte("[")
			valuesCounter = 0
		} else {
			intermediateBuffer = append(intermediateBuffer, values...)
			intermediateBuffer = append(intermediateBuffer, []byte(",")...)
			valuesCounter++
		}
	}
}
