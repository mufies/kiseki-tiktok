package middleware

import (
	"log"
	"net/http"
	"time"
)

type wrapperWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrapperWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &wrapperWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next.ServeHTTP(wrapper, r)
		stop := time.Since(start)

		log.Printf("%s %s -> %d (%v)",
			r.Method,
			r.URL.Path,
			wrapper.statusCode,
			stop,
		)
	})
}
