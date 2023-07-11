package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.klik.doctor/platform/go-pkg/dapr/daprhttp"
	"gitlab.klik.doctor/platform/go-pkg/dapr/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type Response struct {
	CorrelationID string `json:"correlation_id"`
}

func TestLoggerWithTraceparent(t *testing.T) {
	//init new logger
	lg, _ := logger.NewLogger(
		logger.NewGoKitLog(&logger.LogConfig{
			Level: "debug",
		}), "testService",
	)
	ctx := context.Background()
	log := lg.WithContext(ctx)

	ts := httptest.NewServer(otelhttp.NewHandler(http.HandlerFunc(testServer), "server", otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents)))

	defer ts.Close()

	response := &Response{}
	// headers := map[string]string{
	// 	"Authorization": "Bearer lalalala",
	// 	"Content-Type":  "application/json",
	// }
	err := daprhttp.NewRequest(ctx, "POST", ts.URL, nil, response)
	if err != nil {
		log.Error(err)
	}

}

func testServer(w http.ResponseWriter, r *http.Request) {
	//init new logger
	lg, _ := logger.NewLogger(
		logger.NewGoKitLog(&logger.LogConfig{
			Level: "debug",
		}), "testService",
	)

	//request context
	ctx := r.Context()
	log := lg.WithContext(ctx)
	traceparentHeader := r.Header.Get("Traceparent")
	traceID := logger.GetTraceID(ctx)

	//Info Level
	log.Info("starting service, traceparent header  : " + traceparentHeader)

	//inet server B
	ts := httptest.NewServer(otelhttp.NewHandler(http.HandlerFunc(testServerB), "server", otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents)))
	defer ts.Close()

	response := &Response{}
	err := daprhttp.NewRequest(ctx, "GET", ts.URL, nil, response, nil)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"correlation_id": traceID})
	//json.NewEncoder(w).Encode(map[string]string{"correlation_id": traceparentHeader})

}

func testServerB(w http.ResponseWriter, r *http.Request) {
	//init new logger
	lg, _ := logger.NewLogger(
		logger.NewGoKitLog(&logger.LogConfig{
			Level: "info",
		}), "testService",
	)

	ctx := r.Context()
	log := lg.WithContext(ctx)
	traceparentHeader := r.Header.Get("Traceparent")
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	//Info Level
	log.Info("starting serviceB, traceparent header  : " + traceparentHeader)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"correlation_id": traceID})
	//json.NewEncoder(w).Encode(map[string]string{"correlation_id": traceparentHeader})
}
