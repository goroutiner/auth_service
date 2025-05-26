package handlers

import (
	"auth_service/internal/config"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
// если пользователь не активен (временные параметры задаются в congif/config.go).
func СleanupVisitors() error {
	cleanupInterval, err := strconv.Atoi("CLEANUP_INTERVAL")
	if err != nil {
		return fmt.Errorf("env 'CLEANUP_INTERVAL' is not number: %w", err)
	}
	inactivityLimit, err := strconv.Atoi("INACTIVITY_LIMIT")
	if err != nil {
		return fmt.Errorf("env 'INACTIVITY_LIMIT' is not number: %w", err)
	}
	ticker := time.NewTicker(time.Duration(cleanupInterval) * time.Minute)
	for range ticker.C {
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > time.Duration(inactivityLimit)*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
	return nil
}

// getVisitor записывает в словарь visitors лимитеры для заданного ip.
func getVisitor(ip string) (*rate.Limiter, error) {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if exists {
		// Обновляем время последнего запроса пользователя
		v.lastSeen = time.Now()
		return v.limiter, nil
	}

	rateLimit, err := strconv.Atoi(config.RateLimit)
	if err != nil {
		return nil, fmt.Errorf("env 'RATE_LIMIT' is not number: %w", err)
	}
	bufferLimit, err := strconv.Atoi(config.BufferLimit)
	if err != nil {
		return nil, fmt.Errorf("env 'BUFFER_LIMIT' is not number: %w", err)
	}
	v = &visitor{
		limiter:  rate.NewLimiter(rate.Limit(rateLimit), bufferLimit),
		lastSeen: time.Now(),
	}
	visitors[ip] = v
	return v.limiter, nil
}

// LimiterMiddleware проверяет, не превышен ли лимит запросов (RPS) для каждого IP-адреса.
// Если лимит превышен, возвращается ошибка 429 Too Many Requests.
func LimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		limiter, err := getVisitor(ip)
		if err != nil {
			log.Printf("error: limiter middleware is not available: %v\n", err)
		}
		if err == nil && !limiter.Allow() {
			log.Printf("too many requests for user with ip: %s\n", ip)
			http.Error(w, fmt.Sprintf("Too Many Requests for the user: %s", ip), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
