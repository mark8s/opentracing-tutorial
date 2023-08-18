package main

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"log"
	"net/http"
)

func main() {
	log.Println("details service starting ......")
	// Jaeger configuration
	cfg := config.Configuration{
		ServiceName: "details-service",
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
	http.HandleFunc("/details", details)
	err = http.ListenAndServe(":8083", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func details(res http.ResponseWriter, req *http.Request) {
	log.Println("getting details")

	// Get the incoming Span from the context
	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := opentracing.StartSpan("details", opentracing.FollowsFrom(spanCtx))
	defer span.Finish()
	span.SetTag("details", details)

	str := `
床前明月光，
疑是地上霜，
举头望明月，
低头思故乡
`
	res.Write([]byte(str))

}
