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
		tracer.WithService("testrepo"),
		tracer.WithEnv("dev"),
	)
	defer tracer.Stop()

	if err := profiler.Start(
		profiler.WithService("testrepo"),
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
	span, _ := tracer.StartSpanFromContext(req.Context(), "hello_span", tracer.AnalyticsRate(1))
	err := svc(req.Context())
	if err != nil {
		log.Printf("err occured: %v", err)
	}
	fmt.Fprintf(w, "hello\n")
	span.Finish(tracer.WithError(err))
}

func svc(ctx context.Context) error {
	span, _ := tracer.StartSpanFromContext(ctx, "sub_svc_span", tracer.ServiceName("subsvc"), tracer.AnalyticsRate(1))
	time.Sleep(time.Millisecond * 500)
	span.Finish()
	return errors.New("test error")
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
