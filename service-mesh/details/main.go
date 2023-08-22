package main

import (
	"log"
	"net/http"
	"service-mesh/tracing"
)

func main() {
	log.Println("details starting ......")
	_, closer := tracing.Init()
	defer closer.Close()
	http.HandleFunc("/details", details)
	err := http.ListenAndServe(":8083", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func details(res http.ResponseWriter, req *http.Request) {
	log.Println("getting details")

	reqID, _, err := tracing.Extract(req)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("reqID: " + reqID)

	str := `
床前明月光，
疑是地上霜，
举头望明月，
低头思故乡
`
	res.Write([]byte(str))

}
