package main

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"os"
	"sync"

	lackdr "github.com/yringler/go-drop-lack"
	insidescraper "github.com/yringler/inside-chassidus-scraper"
)

type Response struct {
	Source string
}

const dropboxFolder = "inside_chassidus/"
const dropboxFileName = "data.json.gz"
const uploadPath = dropboxFolder + dropboxFileName

func main() {
	mut := sync.Mutex{}
	isFetching := false
	lackdr.AccessToken = os.Getenv("dropbox_token")

	/*
		If the data was already uploaded to dropbox, get a link and sent it back.
		Otherwise, return "not ready".
		Trigger crawl/upload if hasn't been done already.
	*/

	getData := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if link, err := lackdr.GetShareLink(uploadPath); err != nil {
			responseJSON, _ := json.Marshal(Response{
				Source: link,
			})

			w.Write(responseJSON)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)

			if isFetching {
				return
			}

			// Make sure 2 requests don't both trigger scrapes.
			mut.Lock()
			defer mut.Unlock()

			if !isFetching {
				isFetching = true

				// Trigger scarpe/upload.
				go func() {
					scraper := insidescraper.InsideScraper{}

					if err := scraper.Scrape(); err != nil {
						panic(err)
					}

					if err := createDataFile(scraper.Site()); err != nil {
						panic(err)
					}

					if _, err = lackdr.UploadFile(dropboxFileName, dropboxFolder); err != nil {
						panic(err)
					}
				}()
			}
		}
	})

	http.Handle("/", getData)
}

// Crate data file (gzipped new line seperated JSON objects). Return name, and err if error
func createDataFile(site []insidescraper.SiteSection) error {
	var partedJSON string

	for _, value := range site {
		sectionBytes, _ := json.Marshal(value)
		sectionJSON := string(sectionBytes) + "\n"
		partedJSON += sectionJSON
	}

	file, err := os.Create(dropboxFileName)

	if err != nil {
		return err
	}

	defer file.Close()

	zipper := gzip.NewWriter(file)

	if _, err = zipper.Write([]byte(partedJSON)); err != nil {
		return err
	}

	return nil
}
