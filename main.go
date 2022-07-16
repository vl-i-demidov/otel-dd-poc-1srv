package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {
	stopTraceExporter := setUpOtelTracing()
	defer stopTraceExporter()

	stopProfiler := startProfiler()
	defer stopProfiler()

	// Create a traced mux router.
	router := mux.NewRouter()

	router.Use(
		otelmux.Middleware(os.Getenv("DD_SERVICE"), otelmux.WithTracerProvider(otel.GetTracerProvider())),
	)

	// Continue using the router as you normally would.
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		sleep, _ := strconv.Atoi(r.URL.Query().Get("sleep"))
		loadCPU(sleep)
		w.Write([]byte("Hello World!"))
	})
	http.ListenAndServe(":8080", router)
}

func setUpOtelTracing() (stop func()) {

	// create low-level client to export tracing
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(
			"datadog-agent:4317",
		),
		otlptracegrpc.WithDialOption(
			grpc.WithBlock(),
		))

	// create high-level exporting client
	// In real life we should connect in background
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	log.Print("Connecting......")

	traceExp, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected!")
	// tags can be added here
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String(ext.Environment, "qa"),
			attribute.String(ext.ServiceName, "otel-dd-poc-1srv"),
			attribute.String(ext.Version, "0.0.1"),
		))

	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(traceExp),
		tracesdk.WithResource(res),
	)

	// set global propagator to tracecontext
	otel.SetTextMapPropagator(propagation.TraceContext{})
	// set global trace provider
	otel.SetTracerProvider(tracerProvider)

	return func() {
		cxt, cancel := context.WithTimeout(ctx, 50*time.Second)
		defer cancel()

		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}

func startProfiler() (stop func()) {
	err := profiler.Start(
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,

			// The profiles below are disabled by
			// default to keep overhead low, but
			// can be enabled as needed.
			// profiler.BlockProfile,
			// profiler.MutexProfile,
			// profiler.GoroutineProfile,
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		profiler.Stop()
	}
}

func loadCPU(sec int) {
	done := make(chan int)

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
			}
		}()
	}

	time.Sleep(time.Second * time.Duration(sec))
	close(done)
}
