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
	log.Println("summercamp starting ......")

	_, closer := tracing.Init()
	defer closer.Close()

	http.HandleFunc("/summercamp", summerCamp)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func summerCamp(res http.ResponseWriter, req *http.Request) {
	log.Println("welcome to summer camp")

	readingService, ok := os.LookupEnv("READING_SERVICE")
	if !ok {
		readingService = "localhost:8082"
	}

	reqID, spanCtx, err := tracing.Extract(req)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("reqID: " + reqID)
	url := fmt.Sprintf("http://%s/reading", readingService)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tracing.Inject(spanCtx, request, reqID); err != nil {
		log.Println(err)
		return
	}

	resp := makeRequest(request)
	res.Write([]byte(fmt.Sprintf("user: %s, staring summer camp. he are reading poetry\n %s", "", resp)))

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
