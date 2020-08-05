package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	ctx := context.Background()
	redisURL := os.Getenv("REDIS_URL")
	dataURL := os.Getenv("DATA_URL")
	redisOptions, _ := redis.ParseURL(redisURL)
	redisOptions.Username = ""
	rdb := redis.NewClient(redisOptions)

	http.Handle("/check", http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		queryTime := request.URL.Query().Get("date")
		unixTime, err := strconv.ParseInt(queryTime, 10, 64)
		writeBuffer := bytes.Buffer{}

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(writeBuffer.Bytes())
		}

		requesterVersionDate := time.Unix(unixTime, 0)

		currentDate, err := rdb.Get(ctx, "current_date").Time()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(writeBuffer.Bytes())
			return
		}

		if requesterVersionDate.Before(currentDate) {
			http.Redirect(w, request, dataURL, http.StatusPermanentRedirect)
		} else {
			w.WriteHeader(http.StatusNoContent)
			w.Write(writeBuffer.Bytes())
		}
	}))

	password := os.Getenv("AUTH")

	http.Handle("/update", http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		writeBuffer := bytes.Buffer{}

		requestPassword := request.URL.Query().Get("auth")

		if requestPassword != password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(writeBuffer.Bytes())
			return
		}

		queryTime := request.URL.Query().Get("date")
		unixTime, _ := strconv.ParseInt(queryTime, 10, 64)
		requesterVersionDate := time.Unix(unixTime, 0)
		if response := rdb.Set(ctx, "current_date", requesterVersionDate, 0); response.Err() != nil {
			fmt.Println(response.Err())
		}
	}))

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}
