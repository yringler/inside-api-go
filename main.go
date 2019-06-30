package main

import (
	"encoding/json"
	"net/http"
	"sync"

	insidescraper "github.com/yringler/inside-chassidus-scraper"
)

type Response struct {
	Source string
}

const uploadPath = "inside_chassidus/data.json"

func main() {
	mut := sync.Mutex{}
	isFetching := false

	/*
		If the data was already uploaded to dropbox, get a link and sent it back.
		Otherwise, return "not ready".
		Trigger crawl/upload if hasn't been done already.
	*/

	getData := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if link, err := getShareLink(uploadPath); err != nil {
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
					err := scraper.Scrape()

					if err != nil {
						panic(err)
					}

				}()
			}
		}
	})

	http.Handle("/", getData)
}