package crawlerconfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	CrawlersDirectory string `json:"crawlers_directory"`
	URLFilesDirectory string `json:"url_files_directory"`
	RecordsPerFile    int    `json:"records_per_file"`
}

func Load() (*Config, error) {
	c := &Config{}
	file, err := os.Open("config.json")
	if err != nil {
		log.Println("Could not find a valid config file")
		return c, err
	}
	defer file.Close()
	configData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Error while reading the contents")
		return c, err
	}
	err = json.Unmarshal(configData, c)
	if err != nil {
		log.Println("Invalid config detected")
		return c, err
	}
	return c, nil
}

func (c *Config) Save() error {
	file, err := os.Create("config.json")
	if err != nil {
		log.Println("Could not create the config file")
		return err
	}
	defer file.Close()
	configJSON, err := json.Marshal(c)
	if err != nil {
		log.Println("Could not convert config to JSON")
		return err
	}
	writtenBytes, err := file.Write(configJSON)
	if err != nil || writtenBytes != len(configJSON) {
		log.Println("Could write to config file")
		return err
	}
	return nil
}
