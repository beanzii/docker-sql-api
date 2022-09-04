package middlewares

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func LogRequest(w http.ResponseWriter, r *http.Request, next func(http.ResponseWriter, *http.Request)) {
	start := time.Now()
	recorder := StatusRecorder{
		ResponseWriter: w,
	}
	next(&recorder, r)
	ctx := r.Context()
	reqIDRaw := ctx.Value(ContextKeyRequestID)
	reqID, ok := reqIDRaw.(string)
	if !ok {
		log.Printf("[ERROR] Unable to get request id string")
	}
	username, _, _ := r.BasicAuth()
	duration := time.Since(start)
	ri := httpLog{
		StatusCode: recorder.Status,
		Method:     r.Method,
		URL:        r.URL.String(),
		RemoteAddr: r.RemoteAddr,
		Username:   username,
		RequestID:  reqID,
		Duration:   duration,
	}
	logHTTPReq(ri)
}

type httpLog struct {
	StatusCode int
	Duration   time.Duration
	Method     string
	RemoteAddr string
	URL        string
	Username   string
	RequestID  string
}

type StatusRecorder struct {
	Status int
	http.ResponseWriter
}

func (r *StatusRecorder) WritHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func logHTTPReq(ri httpLog) {
	ip := strings.Split(ri.RemoteAddr, ":")
	log.Printf("%s %d %s %s %s %s %v\n", ri.RequestID, ri.StatusCode, ri.Method, ri.URL, ip[0], ri.Username, ri.Duration)
}
