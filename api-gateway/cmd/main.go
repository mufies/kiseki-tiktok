package main

import (
	"log"
	"net/http"
	"strconv"

	"api-gateway/middleware"
	"api-gateway/proxy"
)

func checkHealth(w http.ResponseWriter, r *http.Request) {
	response := []byte(`{"status":"ok"}`)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", checkHealth)
	p := proxy.NewProxy("http://localhost:8081")
	mux.Handle("/api/users/", middleware.AuthMiddleware(p))
	go middleware.CleanupClient()
	log.Fatal(http.ListenAndServe(":8080",
		middleware.RecoveryMiddleware(
			middleware.CORSMiddleware(
				middleware.TimeoutMiddleware(
					middleware.LoggerMiddleware(
						middleware.RateLimitMiddleware(
							middleware.MaxByteMiddleware(mux),
						),
					),
				),
			),
		),
	))
}
