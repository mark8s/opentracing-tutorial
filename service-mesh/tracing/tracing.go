package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
	"io"
	"net/http"
)

// Init returns an instance of Jaeger Tracer.
func Init() (opentracing.Tracer, io.Closer) {
	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	// serviceName, sampler and reporter are not used since we're not emitting traces with this tracer.
	// This tracer is used for extracting and injecting trace header data, to propagate context.
	tracer, closer := jaeger.NewTracer(
		"",
		jaeger.NewConstSampler(false),
		jaeger.NewNullReporter(),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
	)

	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

// Inject injects the outbound HTTP request with the given span's context to ensure
// correct propagation of span context throughout the trace.
func Inject(spanContext opentracing.SpanContext, request *http.Request, requestID string) error {
	request.Header.Add("x-request-id", requestID)
	return opentracing.GlobalTracer().Inject(
		spanContext,
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(request.Header))
}

// Extract extracts the inbound HTTP request to obtain the parent span's context to ensure
// correct propagation of span context throughout the trace.
func Extract(r *http.Request) (string, opentracing.SpanContext, error) {
	requestID := r.Header.Get("x-request-id")
	spanCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	return requestID, spanCtx, err
}
