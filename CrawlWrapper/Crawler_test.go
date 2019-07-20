package crawler

import (
	"log"
	"testing"
)

func TestInvalidFile(t *testing.T) {
	testCrawlingWorker := &Worker{Language: "python", ScriptPath: "hehe.py"}
	data, err := testCrawlingWorker.Crawl("TEST URI")
	if err == nil || len(data) != 0 {
		log.Fatalf("System didn't return an error!")
	}
}
