# Zipkin compatibility features

## `NewZipkinB3HTTPHeaderPropagator()`

Adds support for injecting and extracting Zipkin B3 Propagation HTTP headers,
for use with other Zipkin collectors.

```go

// ...
import (
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/jaegertracing/jaeger-client-go/v3"
	"github.com/jaegertracing/jaeger-client-go/v3/zipkin"
)

func main() {
	// ...

	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	injector := jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, zipkinPropagator)
	extractor := jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, zipkinPropagator)

	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	zipkinSharedRPCSpan := jaeger.TracerOptions.ZipkinSharedRPCSpan(true)

	// create Jaeger tracer
	tracer, closer := jaeger.NewTracer(
		"myService",
		mySampler, // as usual
		myReporter // as usual
		injector,
		extractor,
		zipkinSharedRPCSpan,
	)

	opentracing.SetGlobalTracer(tracer)

    // continue main()
}
```

If you'd like to follow the official guides from https://pkg.go.dev/github.com/jaegertracing/jaeger-client-go/v3/config#example-Configuration-InitGlobalTracer-Production, here is an example.

```go
import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/jaegertracing/jaeger-client-go/v3"
	jaegerClientConfig "github.com/jaegertracing/jaeger-client-go/v3/config"
	"github.com/jaegertracing/jaeger-client-go/v3/zipkin"
	"github.com/jaegertracing/jaeger-client-go/v3/log"
	"github.com/uber/jaeger-lib/metrics"
)

func main(){
	//...
	
	// Recommended configuration for production.
	cfg := jaegercfg.Configuration{}
	
	// Example logger and metrics factory. Use github.com/jaegertracing/jaeger-client-go/v3/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	 
	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	 
	// Create tracer and then initialize global tracer
	closer, err := cfg.InitGlobalTracer(
	  serviceName,
	  jaegercfg.Logger(jLogger),
	  jaegercfg.Metrics(jMetricsFactory),
	  jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
	  jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
	  jaegercfg.ZipkinSharedRPCSpan(true),
	)
	
	if err != nil {
	    log.Printf("Could not initialize jaeger tracer: %s", err.Error())
	    return
	}
	defer closer.Close()
	
	// continue main()
}

```
