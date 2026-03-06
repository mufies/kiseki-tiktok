package middleware

import (
    "net"
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

type client struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

var clients = make(map[string]*client)
var mu sync.Mutex

func getClientLimiter(ip string) *rate.Limiter {
    mu.Lock()
    defer mu.Unlock()
    value, exist := clients[ip]
    if exist {
        value.lastSeen = time.Now()
        return value.limiter
    }
    limiter := rate.NewLimiter(5, 10)
    clients[ip] = &client{
        limiter:  limiter,
        lastSeen: time.Now(),
    }
    return limiter
}

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil {
            ip = r.RemoteAddr
        }
        if !getClientLimiter(ip).Allow() {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func CleanupClient() {
    for {
        time.Sleep(1 * time.Minute) // 👈 thêm cái này, không thì vòng lặp chạy 100% CPU
        mu.Lock()
        for ip, client := range clients {
            if time.Since(client.lastSeen) > 3*time.Minute {
                delete(clients, ip)
            }
        }
        mu.Unlock()
    }
}
