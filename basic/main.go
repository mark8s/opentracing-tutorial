package main

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log2 "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"log"
	"net/http"
	"sync"
)

func main() {
	closer := initJaeger("summer-camp")
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	rootSpan := tracer.StartSpan("ready-for-class")
	rootSpan.SetBaggageItem("userid", "mark")
	rootSpan.SetBaggageItem("traffic", "staging")
	defer rootSpan.Finish()
	spanCtx := opentracing.ContextWithSpan(context.Background(), rootSpan)

	log.Println("welcome to summer camp")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		http.HandleFunc("/reading", func(w http.ResponseWriter, r *http.Request) {
			// traceid 对应的header为 Uber-Trace-Id
			// 上文中设置的baggage userid header标 变为 Uberctx-Userid
			// 上文中设置的baggage traffic header标 变为 Uberctx-traffic
			spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header)) // r.Header 中包含了 Uber-Trace-Id: 根traceId
			span := tracer.StartSpan("reading", ext.RPCServerOption(spanCtx))
			defer span.Finish()
			span.SetTag("course", "reading")
			str := `
床前明月光，
疑是地上霜，
举头望明月，
低头思故乡
`
			span.LogFields(
				log2.String("event", "reading 靜夜思李白"),
				log2.String("user", span.BaggageItem("userid")),
			)
			w.Write([]byte(str))
			// 调用writing
			writing(opentracing.ContextWithSpan(context.Background(), span))
		})

		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	go func() {
		defer wg.Done()
		http.HandleFunc("/writing", func(w http.ResponseWriter, r *http.Request) {
			spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header)) // r.Header 中包含了 Uber-Trace-Id: 根traceId
			span := tracer.StartSpan("writing", ext.RPCServerOption(spanCtx))
			defer span.Finish()
			span.SetTag("course", "writing")

			str := `
Exploring Ancient Ruins

Last week, I visited ancient ruins with my class. The massive stones and intricate carvings were fascinating. 
I imagined what life was like for people who lived there long ago. It was like stepping into a time machine. 
I hope to learn more about history and explore more sites like this.
`
			span.LogFields(
				log2.String("event", "writing english essay"),
				log2.String("user", span.BaggageItem("userid")),
			)
			w.Write([]byte(str))
		})

		log.Fatal(http.ListenAndServe(":8082", nil))
	}()

	reading(spanCtx)
	wg.Done()
}

func initJaeger(service string) io.Closer {
	conf := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: "http://10.10.13.44:14268/api/traces",
		},
	}

	closer, err := conf.InitGlobalTracer(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return closer
}

func reading(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "reading")
	defer span.Finish()
	method := "GET"
	url := "http://localhost:8081/reading"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot request reading: %v\n", err))
	}

	body, err := io.ReadAll(resp.Body)
	span.LogFields(
		log2.String("event", "reading"),
		log2.String("value", string(body)),
		log2.String("user", span.BaggageItem("userid")),
	)
}

func writing(ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "writing")
	defer span.Finish()
	method := "GET"
	url := "http://localhost:8081/writing"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot request writing: %v\n", err))
	}

	body, err := io.ReadAll(resp.Body)
	span.LogFields(
		log2.String("event", "writing"),
		log2.String("value", string(body)),
		log2.String("user", span.BaggageItem("userid")),
	)
}
