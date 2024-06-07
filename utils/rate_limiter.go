package utils

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter   *rate.Limiter
	last_seen time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.RWMutex
)

func getClient(ip string) *rate.Limiter {
	reqPerSecString := LoadEnv("RATE_LIMIT_REQ_PER_SEC")
	reqPerSec, err := strconv.Atoi(reqPerSecString)
	if err != nil {
		log.Fatal("Error converting RATE_LIMIT_REQ_PER_SEC to integer")
	}

	mu.RLock()
	user, exists := clients[ip]
	mu.RUnlock()
	if !exists {
		mu.Lock()
		limiter := rate.NewLimiter(1, reqPerSec)
		clients[ip] = &client{limiter, time.Now()}
		mu.Unlock()
		return limiter
	}
	user.last_seen = time.Now()
	return user.limiter
}

func CleanupUsers() {
	for {
		time.Sleep(time.Minute)
		mu.RLock()
		for ip, v := range clients {
			if time.Since(v.last_seen) < 2*time.Minute {
				delete(clients, ip)
			}
		}
		mu.RUnlock()
	}
}

func Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		limiter := getClient(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
