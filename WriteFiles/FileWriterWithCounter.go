/*
Package filewriterwithcounter writes data to a file and maintains counter of written files
*/
package filewriterwithcounter

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type fileWriterWithCounter struct {
	dataDirectory string
}

var instance *fileWriterWithCounter
var once sync.Once

func createDataDirectory() string {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		log.Fatalln("Could not read fs")
	}
	dataDirectoryPath := path.Join(currentWorkingDir, "Data")
	_, err = os.Stat(dataDirectoryPath)
	if err != nil {
		log.Println("Data directory doesn't exist, creating new directory")
		err = os.Mkdir(dataDirectoryPath, os.ModePerm)
		if err != nil {
			log.Fatalln("Could not create data directory")
		}
	}
	log.Printf("Data directory ready! (%s)\n", dataDirectoryPath)
	return dataDirectoryPath
}

//GetInstance returns pointer to the instance of fileWriterWithCounter
func GetInstance() *fileWriterWithCounter {
	once.Do(func() {
		dataDirectory := createDataDirectory()
		instance = &fileWriterWithCounter{dataDirectory: dataDirectory}
	})
	return instance
}

func (fwc *fileWriterWithCounter) Write(ch <-chan []byte, crawlerName string) {
	sequenceNumber := 0
	for crawledDataChunk := range ch {
		sequenceNumber++
		log.Println("Received new batch, saving to file...")
		filename := fmt.Sprintf("%s-%d-%s.json", crawlerName, sequenceNumber, time.Now().Format("03-04-05-pm-02-01-2006"))
		filePath := path.Join(fwc.dataDirectory, filename)
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Println(err)
			log.Fatalln("Failed to create data file")
		}
		n, err := file.Write(crawledDataChunk)
		file.Close()
		if err != nil {
			log.Fatalln("Failed to write data")
		} else if n != len(crawledDataChunk) {
			log.Fatalln("Failed to write data. Partial data written to file.")
		}
	}
}
