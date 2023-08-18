package main

import (
	"context"
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
	log.Println("summercamp service starting ......")

	// Jaeger configuration
	cfg := config.Configuration{
		ServiceName: "summer-camp-service",
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
	http.HandleFunc("/summercamp", summerCamp)
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func summerCamp(res http.ResponseWriter, req *http.Request) {
	log.Println("welcome to summer camp")

	ctx := opentracing.ContextWithSpan(context.Background(), opentracing.GlobalTracer().StartSpan("summer-camp"))

	span := opentracing.SpanFromContext(ctx)
	defer span.Finish()

	user := span.BaggageItem("userid")
	span.LogKV("event", "come in summer camp", "user", user)

	request, err := http.NewRequestWithContext(ctx, "GET", "http://reading-service:8082/reading", nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot request writing: %v\n", err))
	}

	body, err := io.ReadAll(resp.Body)
	res.Write([]byte(fmt.Sprintf("user: %s, staring summer camp. he are reading poetry\n %s", user, body)))

}
