package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
	rules := []tracer.SamplingRule{tracer.RateRule(1)}
	tracer.Start(
		tracer.WithSamplingRules(rules),
		tracer.WithService("goserver"),
		tracer.WithEnv("dev"),
	)
	defer tracer.Stop()

	if err := profiler.Start(
		profiler.WithService("goserver"),
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
	http.ListenAndServe(":80", mux)
}

func hello(w http.ResponseWriter, req *http.Request) {
	span, _ := tracer.StartSpanFromContext(req.Context(), "hello_span", tracer.AnalyticsRate(1))
	err := svc()
	if err != nil {
		log.Printf("err occured: %v", err)
	}
	fmt.Fprintf(w, "hello\n")
	span.Finish(tracer.WithError(err))
}

func svc() error {
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
