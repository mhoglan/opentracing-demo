package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"

	"github.com/TykTechnologies/tyk/trace"
	"github.com/opentracing/opentracing-go/ext"
	// "github.com/TykTechnologies/tyk/trace"
)

type mainHandler struct {
	h http.Handler
}

const service = "trace-service"

var destination = os.Getenv("DEST_URL")
var appid = os.Getenv("APP_ID")
var appPort = os.Getenv("APP_PORT")
var tracer = os.Getenv("APP_TRACER")

var serviceName = appid

func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lg.Info(r.Method, " ==> ", r.URL.Path)
	span, req := trace.Root(serviceName, r)
	defer span.Finish()
	m.h.ServeHTTP(w, req)

}

const zipkinConfig = `{
    "reporter": {
        "url": "http://localhost:9411/api/v2/spans"
    }
}`

const JaegerConfig = `{
	"serviceName": "trace-app",
	"disabled": false,
	"rpc_metrics": false,
	"tags": null,
	"sampler": {
		"type": "const",
		"param": 1,
		"samplingServerURL": "",
		"maxOperations": 0,
		"samplingRefreshInterval": 0
	},
	"reporter": {
		"queueSize": 0,
		"BufferFlushInterval": 0,
		"logSpans": true,
		"localAgentHostPort": "jaeger:6831",
		"collectorEndpoint": "",
		"user": "",
		"password": ""
	},
	"headers": null,
	"baggage_restrictions": null,
	"throttler": null
}`

type logger struct{}

func (logger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] %s\n", fmt.Sprintf(format, args...))
}
func (logger) Info(args ...interface{}) {
	fmt.Printf("[INFO] %s\n", fmt.Sprint(args...))
}

func (logger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO] %s\n", fmt.Sprintf(format, args...))
}

var lg = logger{}

func main() {
	port := 6666
	if appPort != "" {
		p, err := strconv.Atoi(appPort)
		if err != nil {
			log.Fatal(err)
		}
		port = p
	}
	var jsonConfig string
	switch tracer {
	case "", "jaeger":
		jsonConfig = JaegerConfig
		if tracer == "" {
			tracer = "jaeger"
		}
	case "zipkin":
		jsonConfig = zipkinConfig
	}
	var o map[string]interface{}
	err := json.Unmarshal([]byte(jsonConfig), &o)
	if err != nil {
		log.Fatal(err)
	}

	trace.SetupTracing(tracer, o)
	trace.SetLogger(lg)
	defer trace.Close()
	err = trace.AddTracer(serviceName)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", ping)
	mux.HandleFunc("/pong", pong)
	mux.HandleFunc("/echo", echo)
	h := mainHandler{h: mux}
	log.Println("starting trace service at :", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), h))
}

func ping(w http.ResponseWriter, r *http.Request) {
	span, ctx := trace.Span(r.Context(), "ping")
	defer span.Finish()
	if destination != "" {
		req, err := http.NewRequest(http.MethodGet, destination+"/pong", nil)
		if err != nil {
			serveErr(w, err)
			return
		}
		res, err := do(ctx, "call-pong", req)
		if err != nil {
			serveErr(w, err)
			return
		}
		defer res.Body.Close()
		io.Copy(w, res.Body)
	}
}

func serveErr(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}

func do(ctx context.Context, name string, req *http.Request) (*http.Response, error) {
	span, ctx := trace.Span(ctx, name)
	defer span.Finish()
	span.SetTag(
		ext.SpanKindRPCClient.Key,
		ext.SpanKindRPCClient.Value,
	)
	span.SetTag(
		string(ext.HTTPUrl),
		req.URL.String(),
	)
	span.SetTag(
		string(ext.HTTPMethod),
		req.Method,
	)
	err := trace.InjectFromContext(ctx, span, req.Header)
	if err != nil {
		lg.Errorf(err.Error())
	}
	return http.DefaultClient.Do(req)
}

func pong(w http.ResponseWriter, r *http.Request) {
	span, ctx := trace.Span(r.Context(), "pong")
	defer span.Finish()
	if destination != "" {
		req, err := http.NewRequest(http.MethodGet, destination+"/echo", nil)
		if err != nil {
			serveErr(w, err)
			return
		}
		res, err := do(ctx, "call-echo", req)
		if err != nil {
			serveErr(w, err)
			return
		}
		defer res.Body.Close()
		io.Copy(w, res.Body)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	span, _ := trace.Span(r.Context(), "echo")
	defer span.Finish()
	b, _ := httputil.DumpRequest(r, true)
	w.Write(b)
}
