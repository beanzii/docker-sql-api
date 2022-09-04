package middlewares

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const ContextKeyRequestID ContextKey = "requestID"

func ReqIdRequest(w http.ResponseWriter, r *http.Request, next func(http.ResponseWriter, *http.Request)) {
	ctx := r.Context()
	id := uuid.New()
	log.Printf("[INFO] Incoming request: %s % s %s %s", r.Method, r.RequestURI, r.RemoteAddr, id.String())
	ctx = context.WithValue(ctx, ContextKeyRequestID, id.String())
	r = r.WithContext(ctx)
	next(w, r)
}
