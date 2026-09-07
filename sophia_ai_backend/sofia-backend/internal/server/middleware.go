package server

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	authctx "github.com/armandoalvarado/sofia-backend/internal/auth/context"
	authjwt "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure/jwt"
	"github.com/armandoalvarado/sofia-backend/internal/platform/httpjson"
)

type Middleware func(http.Handler) http.Handler

type RateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	requests map[string]rateBucket
}

type rateBucket struct {
	count      int
	windowEnd  time.Time
	lastAccess time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 10
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{limit: limit, window: window, now: time.Now, requests: map[string]rateBucket{}}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	if l == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(rateKey(r)) {
			httpjson.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	bucket := l.requests[key]
	if bucket.windowEnd.IsZero() || !bucket.windowEnd.After(now) {
		bucket = rateBucket{windowEnd: now.Add(l.window)}
	}
	bucket.count++
	bucket.lastAccess = now
	l.requests[key] = bucket
	l.prune(now)
	return bucket.count <= l.limit
}

func (l *RateLimiter) prune(now time.Time) {
	for key, bucket := range l.requests {
		if bucket.windowEnd.Before(now.Add(-l.window)) || bucket.lastAccess.Before(now.Add(-2*l.window)) {
			delete(l.requests, key)
		}
	}
}

func rateKey(r *http.Request) string {
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if host != "" {
		if idx := strings.Index(host, ","); idx >= 0 {
			host = host[:idx]
		}
		return strings.TrimSpace(host) + ":" + r.URL.Path
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || strings.TrimSpace(host) == "" {
		host = r.RemoteAddr
	}
	return strings.TrimSpace(host) + ":" + r.URL.Path
}

func AuthMiddleware(tokenService *authjwt.Service) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				httpjson.Unauthorized(w)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				httpjson.Unauthorized(w)
				return
			}

			claims, err := tokenService.Validate(strings.TrimSpace(parts[1]))
			if err != nil {
				httpjson.Unauthorized(w)
				return
			}

			ctx := authctx.WithClaims(r.Context(), authctx.Claims{
				UserID: claims.UserID,
				Email:  claims.Email,
				Role:   claims.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v", recovered)
				httpjson.InternalServerError(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func failedRequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		if recorder.status >= http.StatusBadRequest {
			log.Printf("request failed method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, recorder.status, time.Since(started))
		}
	})
}

func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(allowed) > 0 {
			origin := r.Header.Get("Origin")
			if _, ok := allowed[origin]; ok || allowsWildcard(allowed) {
				allowedOrigin := origin
				if allowedOrigin == "" && allowsWildcard(allowed) {
					allowedOrigin = "*"
				}
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowsWildcard(allowed map[string]struct{}) bool {
	_, ok := allowed["*"]
	return ok
}
