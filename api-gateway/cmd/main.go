package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"api-gateway/middleware"
	"api-gateway/proxy"
)

type ServiceHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

type HealthResponse struct {
	Status   string          `json:"status"`
	Services []ServiceHealth `json:"services"`
	CheckedAt string         `json:"checked_at"`
}

func checkServiceHealth(url string, timeout time.Duration) string {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return "down"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "up"
	}
	return "degraded"
}

func checkHealth(w http.ResponseWriter, r *http.Request) {
	services := []struct {
		name string
		url  string
	}{
		{"User Service", "http://localhost:8081/health"},
		{"Video Service", "http://localhost:8082/health"},
		{"Interaction Service", "http://localhost:8084/actuator/health"},
		{"Event Service", "http://localhost:5001/health"},
		{"Feed Service", "http://localhost:8001/health"},
	}

	results := make([]ServiceHealth, len(services))
	var wg sync.WaitGroup

	for i, svc := range services {
		wg.Add(1)
		go func(idx int, service struct {
			name string
			url  string
		}) {
			defer wg.Done()
			status := checkServiceHealth(service.url, 2*time.Second)
			results[idx] = ServiceHealth{
				Name:   service.name,
				Status: status,
				URL:    service.url,
			}
		}(i, svc)
	}

	wg.Wait()

	overallStatus := "healthy"
	for _, result := range results {
		if result.Status == "down" {
			overallStatus = "unhealthy"
			break
		} else if result.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	healthResp := HealthResponse{
		Status:    overallStatus,
		Services:  results,
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(healthResp); err != nil {
		log.Printf("Error encoding health response: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", checkHealth)

	// User Service proxy - Authentication endpoints (no auth required)
	userServiceProxy := proxy.NewProxy("http://localhost:8081")
	mux.Handle("/auth/", userServiceProxy)

	// User Service proxy - Protected user endpoints
	mux.Handle("/api/users/", middleware.AuthMiddleware(userServiceProxy))

	// Video Service proxy
	videoProxy := proxy.NewProxy("http://localhost:8082")
	mux.Handle("/api/videos/", http.StripPrefix("/api", videoProxy))

	// Interaction Service proxy
	interactionProxy := proxy.NewProxy("http://localhost:8084")
	mux.Handle("/interactions/", interactionProxy)

	// Event Service proxy
	eventProxy := proxy.NewProxy("http://localhost:5001")
	mux.Handle("/events/", eventProxy)
	mux.Handle("/profile/", eventProxy)

	// Feed Service proxy
	feedProxy := proxy.NewProxy("http://localhost:8001")
	mux.Handle("/feed/", feedProxy)
	mux.Handle("/trending", feedProxy)

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
