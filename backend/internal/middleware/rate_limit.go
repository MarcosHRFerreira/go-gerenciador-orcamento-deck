package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count      int
	windowEnds time.Time
}

type rateLimiter struct {
	entries map[string]rateLimitEntry
	limit   int
	window  time.Duration
	mutex   sync.Mutex
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		entries: make(map[string]rateLimitEntry),
		limit:   limit,
		window:  window,
	}
}

func (l *rateLimiter) allow(key string, now time.Time) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.cleanup(now)

	entry, exists := l.entries[key]
	if !exists || now.After(entry.windowEnds) {
		l.entries[key] = rateLimitEntry{
			count:      1,
			windowEnds: now.Add(l.window),
		}
		return true
	}

	if entry.count >= l.limit {
		return false
	}

	entry.count++
	l.entries[key] = entry
	return true
}

func (l *rateLimiter) cleanup(now time.Time) {
	for key, entry := range l.entries {
		if now.After(entry.windowEnds) {
			delete(l.entries, key)
		}
	}
}

func RateLimit(limit int, window time.Duration, scope string) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	return func(c *gin.Context) {
		now := time.Now()
		key := strings.TrimSpace(scope) + ":" + clientIP(c)
		if limiter.allow(key, now) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"message": "too many requests",
		})
	}
}

func clientIP(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	if ip == "" {
		return "unknown"
	}

	return ip
}
