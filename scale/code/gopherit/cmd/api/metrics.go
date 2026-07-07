package main

import (
	"expvar"
	"net/http"
	"strconv"
	"time"
)

// Observability, chapter 10. We keep it dependency-free with the standard
// library's expvar, which publishes variables as JSON at /debug/vars. In
// production you'd usually export Prometheus metrics instead (the chapter shows
// how), but the shape is the same: count requests, time them, expose them.
var (
	metricRequests    = expvar.NewInt("requests_total")
	metricErrors      = expvar.NewInt("responses_5xx_total")
	metricInFlight    = expvar.NewInt("requests_in_flight")
	metricTotalMicros = expvar.NewInt("request_duration_micros_total")
)

// statusRecorder remembers the status code so the metrics layer can see it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// metrics records one measurement per request: count, in-flight gauge,
// duration, and 5xx errors.
func (app *application) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metricRequests.Add(1)
		metricInFlight.Add(1)
		defer metricInFlight.Add(-1)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		metricTotalMicros.Add(time.Since(start).Microseconds())
		if rec.status >= 500 {
			metricErrors.Add(1)
		}
	})
}

// metricsHandler serves a compact plaintext summary at /metrics — the golden
// signals (traffic, errors, latency, saturation) at a glance.
func (app *application) metricsHandler(w http.ResponseWriter, r *http.Request) {
	total := metricRequests.Value()
	avg := int64(0)
	if total > 0 {
		avg = metricTotalMicros.Value() / total
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(
		"requests_total " + strconv.FormatInt(total, 10) + "\n" +
			"responses_5xx_total " + strconv.FormatInt(metricErrors.Value(), 10) + "\n" +
			"requests_in_flight " + strconv.FormatInt(metricInFlight.Value(), 10) + "\n" +
			"request_duration_micros_avg " + strconv.FormatInt(avg, 10) + "\n"))
}
