//Package urlprovider watches the filesystem for changes and parses files with URLs for crawling
package urlprovider

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/fsnotify/fsnotify"
)

var innerURLChannel chan string

//Init initializes the watcher service
func Init(inputDirectoryPath string) {
	log.Printf("Started watching directory '%s' for changes...", inputDirectoryPath)
	innerURLChannel = make(chan string)
	go watchFileSystem(inputDirectoryPath)
}

//GetUrls takes a channel of strings and pushes parsed URLs from file to the CrawlerWrapper
func GetUrls(globalURLChannel chan<- string) {
	for url := range innerURLChannel {
		globalURLChannel <- url
	}
}

func readFileAndExtractUrls(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalln("Error while reading the file")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		go func(url string) {
			innerURLChannel <- url
		}(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		markFileWithStatus(filepath, "errored")
		log.Fatal(err)
	} else {
		markFileWithStatus(filepath, "done")
	}
}

func markFileWithStatus(filename, status string) {
	log.Printf("Done processing '%s'. Marking %s.", filename, status)
	os.Rename(filename, fmt.Sprintf("%s.%s", filename, status))
}

func watchFileSystem(inputDirectoryPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Could not create a filesystem watcher")
	}

	err = watcher.Add(inputDirectoryPath)
	if err != nil {
		log.Fatalln("Failed to watch the directory", inputDirectoryPath)
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Op == 1 {
				matched, err := regexp.MatchString(`(.*)\.(done|errored)$`, event.Name)
				if err != nil {
					fmt.Println("Could not process file:", event.Name)
				}
				if !matched {
					log.Println("New file detected:", event.Name)
					log.Println("Starting processing...")
					go readFileAndExtractUrls(event.Name)
				}
			}
		case err := <-watcher.Errors:
			log.Println(err)
		}
	}
}
