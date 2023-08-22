package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"service-mesh/tracing"
)

func main() {
	log.Println("reading starting ......")
	_, closer := tracing.Init()
	defer closer.Close()
	http.HandleFunc("/reading", reading)
	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func reading(res http.ResponseWriter, req *http.Request) {
	log.Println("start reading")

	detailsService, ok := os.LookupEnv("DETAILS_SERVICE")
	if !ok {
		detailsService = "localhost:8083"
	}

	reqID, spanCtx, err := tracing.Extract(req)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("reqID: " + reqID)
	poetry := "静夜思（唐'李白）"
	url := fmt.Sprintf("http://%s/details", detailsService)
	request, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Inject trace headers into the request in order for Istio to correlate outbound with inbound calls.
	if err := tracing.Inject(spanCtx, request, reqID); err != nil {
		return
	}

	resp := makeRequest(request)
	message := fmt.Sprintf("reading poetry: %s\nDetails: %s", poetry, resp)
	res.Write([]byte(message))

}

func makeRequest(req *http.Request) string {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return ""
	}

	return fmt.Sprintf("%s\n", string(body))
}
