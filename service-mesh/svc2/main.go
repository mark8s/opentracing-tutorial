package main

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"io"
	"log"
	"net/http"
)

func main() {
	log.Println("reading service starting ......")

	// Jaeger configuration
	cfg := config.Configuration{
		ServiceName: "reading-service",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: "http://jaeger.service-mesh:14268/api/traces", // Your Jaeger Collector URL
		},
	}

	tracer, closer, err := cfg.NewTracer(
		config.Logger(jaeger.StdLogger),
		config.Metrics(metrics.NullFactory),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Jaeger: %v", err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)
	http.HandleFunc("/reading", reading)
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func reading(res http.ResponseWriter, req *http.Request) {
	log.Println("start reading")

	// Get the incoming Span from the context
	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := opentracing.StartSpan("reading", opentracing.FollowsFrom(spanCtx))
	defer span.Finish()

	poetry := "静夜思（唐'李白）"
	span.SetTag("poetry", poetry)

	request, err := http.NewRequestWithContext(req.Context(), "GET", "http://details-service:8083/details", nil)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body := "No details found"
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		body = string(bodyBytes)
	}

	message := fmt.Sprintf("reading poetry: %s\nDetails: %s", poetry, body)
	res.Write([]byte(message))

}
