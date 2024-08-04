package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

var defaultSleepTimeMs *int
var defaultStatusCode *int

func SleepServer(w http.ResponseWriter, r *http.Request) {
	sleep, err := time.ParseDuration(r.URL.Query().Get("sleep"))
	if err != nil {
		sleep = time.Duration(*defaultSleepTimeMs) * time.Millisecond
	}
	time.Sleep(sleep)
	w.WriteHeader(*defaultStatusCode)
}

func main() {
	port := flag.String("port", "6000", "port to serve on")
	defaultSleepTimeMs = flag.Int("sleep", 1000, "default sleep time")
	defaultStatusCode = flag.Int("status", 200, "default status code")
	flag.Parse()
	fmt.Printf("Starting server at port " + *port + "\n")

	http.HandleFunc("/", SleepServer)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		fmt.Println(err)
	}
}
