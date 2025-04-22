package handlers

import (
	"auth_service/internal/config"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	visitors = make(map[string]*visitor) // visitors словарь для связи ip -> visitor
	mu       sync.Mutex
)

// visitor внутренняя структура для хранения лимитера и времени последнего запроса.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// СleanupVisitors очищает словарь visitors через каждый временной интервал,
// если пользователь не активен (временные параметры задаются в congif/setting.go).
func СleanupVisitors() {
	ticker := time.NewTicker(config.CleanupInterval)
	for range ticker.C {
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > config.InactivityLimit {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// getVisitor записывает в словарь visitors лимитеры для заданного ip.
func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if exists {
		// Обновляем время последнего запроса пользователя
		v.lastSeen = time.Now()
		return v.limiter
	}

	v = &visitor{
		limiter:  rate.NewLimiter(config.RateLimit, config.BufferLimit),
		lastSeen: time.Now(),
	}
	visitors[ip] = v
	return v.limiter
}

// LimiterMiddleware проверяет, не превышен ли лимит запросов (RPS) для каждого IP-адреса.
// Если лимит превышен, возвращается ошибка 429 Too Many Requests.
func LimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		limiter := getVisitor(ip)
		if !limiter.Allow() {
			log.Println("Too Many Requests")
			http.Error(w, fmt.Sprintf("Too Many Requests for the user: %s", ip), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
