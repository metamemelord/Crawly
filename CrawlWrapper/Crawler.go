//Package crawler wraps an underlying crawling worker and provides a high level API
package crawler

import (
	"os"
	"os/exec"
)

//Crawler interface provides Crawl function to accept a URL and return crawled data and error
type Crawler interface {
	Crawl(string) ([]byte, error)
}

//Worker type represents a crawler with language and path to script
type Worker struct {
	Language   string
	ScriptPath string
}

//Crawl function enables Crawler interface for type *Worker
func (w *Worker) Crawl(url string) ([]byte, error) {
	crawlWorker := exec.Command(w.Language, w.ScriptPath, url)
	crawlWorker.Stderr = os.Stderr
	result, err := crawlWorker.Output()
	return result, err
}
