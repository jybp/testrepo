package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
	rules := []tracer.SamplingRule{tracer.RateRule(1)}
	tracer.Start(
		tracer.WithSamplingRules(rules),
		tracer.WithService("testrepo2"),
		tracer.WithEnv("dev"),
	)
	defer tracer.Stop()

	if err := profiler.Start(
		profiler.WithService("testrepo2"),
		profiler.WithEnv("dev"),
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
	); err != nil {
		log.Fatal(err)
	}
	defer profiler.Stop()

	// Create a traced mux router
	mux := httptrace.NewServeMux()
	// Continue using the router as you normally would.
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/error", serveFile)
	http.ListenAndServe(":8080", mux)
}

func hello(w http.ResponseWriter, req *http.Request) {
	span, ctx := tracer.StartSpanFromContext(req.Context(), "hello_span", tracer.AnalyticsRate(1))
	err := svc(ctx)
	if err != nil {
		log.Printf("err occured: %v", err)
	}
	fmt.Fprintf(w, "hello\n")
	span.Finish(tracer.WithError(err))
}

func svc(ctx context.Context) error {
	span, ctx := tracer.StartSpanFromContext(ctx, "svc2", tracer.ServiceName("svc2"), tracer.AnalyticsRate(1))
	time.Sleep(time.Millisecond * 500)
	err := subsvc(ctx)
	span.Finish(tracer.WithError(err))
	return err
}

func subsvc(ctx context.Context) error {
	span1, ctx := tracer.StartSpanFromContext(ctx, "subsvc2", tracer.ServiceName("subsvc2"), tracer.AnalyticsRate(1))
	time.Sleep(time.Millisecond * 500)
	span1.Finish()

	span2, ctx := tracer.StartSpanFromContext(ctx, "subsvc3", tracer.ServiceName("subsvc3"), tracer.AnalyticsRate(1))
	err := errors.New("test error")
	time.Sleep(time.Millisecond * 500)
	span2.Finish(tracer.WithError(err))
	return err
}

func serveFile(w http.ResponseWriter, req *http.Request) {
	file, err := getFile()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "Filename: %s\n", file)
}

func getFile() (string, error) {
	return "", errors.New("no valid file")
}
