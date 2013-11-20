package main

import (
	"encoding/json"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Message struct {
	Message string `json:"message"`
	Delay   int64  `json:"delay"`
	Path    string `json:"path"`
}

type Url struct {
	Url string `json:"url"`
}

type UrlList struct {
	Urls []Url `json:"urls"`
}

func pickDelayTime(delayStr string) (delay time.Duration) {
	seconds, err := strconv.Atoi(delayStr)
	if err != nil && delayStr != "" {
		log.Printf("Invalid delay argument '%s': %s", delayStr, err)
		delay = time.Duration(rand.Int63n(10))
	} else if seconds < 0 {
		delay = time.Duration(rand.Int63n(10))
	} else {
		delay = time.Duration(seconds)
	}

	delay *= time.Second
	return delay
}

func pickStatusCode(codeStr string) (code int) {
	// Use the code we got passed in
	var err error
	if codeStr != "" {
		code, err = strconv.Atoi(codeStr)
		if err != nil && codeStr != "" {
			log.Printf("Invalid code argument '%s': %s", codeStr, err)
			code = 200
		}

		// Pick a random code
	} else {
		rnd := rand.Intn(10)
		if rnd == 4 {
			code = 400
		} else if rnd == 5 {
			code = 500
		} else {
			code = 200
		}
	}

	return code
}

// Params:
// * code: status code to return (optional). If not provided, status code will
//         be generated randomly
// * delay: number of seconds to wait before returning response. Will be random
// 		    if not provided.
//
// TODO: check Accept header and return JSON or text/plain (URL per line)
func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.String())

	// get the number of seconds to wait to respond
	delay := pickDelayTime(r.URL.Query().Get("delay"))

	log.Printf("Going to wait %d seconds...\n", delay)
	time.Sleep(delay)

	message := Message{Delay: int64(delay.Seconds()), Path: r.URL.Path}

	code := pickStatusCode(r.URL.Query().Get("code"))

	if code >= 500 {
		message.Message = "server error"
	} else if code >= 400 {
		message.Message = "client error"
	} else {
		message.Message = "success"
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshalling to json!")
		fmt.Println(err)
		jsonMessage = []byte("{\"message\":\"Error marshalling json\"}")
	}

	w.Header().Set("Content-Type", "application/json")
	log.Printf("Sending %d: %s", code, jsonMessage)

	w.WriteHeader(code)
	fmt.Fprintf(w, "%s", jsonMessage)
}

func sampleHandler(w http.ResponseWriter, r *http.Request) {
	numUrlsString := r.URL.Query().Get("n")
	numUrls := 100
	if numUrlsString != "" {
		numUrls, _ = strconv.Atoi(numUrlsString)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	urlList := UrlList{}
	for i := 0; i < numUrls; i++ {
		u, err := uuid.NewV4()
		if err != nil {
			fmt.Println("Error with uuid")
			u = nil
		}
		url := Url{Url: fmt.Sprintf("http://%s/%s", r.Host, u)}
		urlList.Urls = append(urlList.Urls, url)
	}
	jsonMessage, _ := json.Marshal(urlList)
	fmt.Fprintf(w, "%s", jsonMessage)
}

func main() {
	http.HandleFunc("/", handler)             // redirect all urls to the handler function
	http.HandleFunc("/sample", sampleHandler) // get a list of URLs

	log.Printf("Listening on ':%s'", os.Getenv("PORT"))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil) // listen for connections at port 9999 on the local machine
	if err != nil {
		log.Printf("Failed to start server: %s", err)
	}
}
