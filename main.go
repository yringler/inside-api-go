package main

import (
	"bytes"
	"context"
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
		requestersUnix, err := strconv.ParseInt(queryTime, 10, 64)
		writeBuffer := bytes.Buffer{}

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeBuffer.WriteString(err.Error())
			w.Write(writeBuffer.Bytes())
			return
		}

		redisUnix, err := rdb.Get(ctx, "current_date").Int64()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeBuffer.WriteString(err.Error())
			w.Write(writeBuffer.Bytes())
			return
		}

		if redisUnix > requestersUnix {
			http.Redirect(w, request, dataURL, http.StatusTemporaryRedirect)
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
		newVersionDate := time.Unix(unixTime, 0)
		if response := rdb.Set(ctx, "current_date", newVersionDate.Unix(), 0); response.Err() != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeBuffer.WriteString(response.Err().Error())
		}

		w.Write(writeBuffer.Bytes())
	}))

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}
