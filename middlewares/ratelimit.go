package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Budgets for the unauthenticated endpoints. These are meant to be invisible
// to a real client and expensive for a script.
const (
	// Total /login requests from one IP, whether they succeed or not.
	LoginIPMax    = 20
	LoginIPWindow = 5 * time.Minute

	// Failed /login attempts against one email address, counted across all
	// IPs so a distributed guesser cannot dodge the per-IP budget.
	LoginAccountMax    = 5
	LoginAccountWindow = 15 * time.Minute

	// Account creation is deliberately tight: each call costs a bcrypt hash
	// and a file write.
	RegisterMax    = 5
	RegisterWindow = time.Hour

	// /refresh and /logout are cheap but still write to the database.
	SessionMax    = 30
	SessionWindow = 5 * time.Minute
)

// counter is one fixed window of hits against a single key.
type counter struct {
	hits     int
	expireAt time.Time
}

// Limiter is a fixed-window counter keyed by an arbitrary string. Expired
// windows are pruned in the background so the map cannot grow without bound
// as new IPs and email addresses arrive.
type Limiter struct {
	mu   sync.Mutex
	max  int
	per  time.Duration
	keys map[string]*counter
}

func NewLimiter(max int, per time.Duration) *Limiter {
	l := &Limiter{
		max:  max,
		per:  per,
		keys: make(map[string]*counter),
	}

	go l.pruneLoop()

	return l
}

// Consume records one hit against key. It reports whether that hit was within
// budget and, when it was not, how long the caller has to wait.
func (l *Limiter) Consume(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	c, ok := l.keys[key]
	if !ok || now.After(c.expireAt) {
		l.keys[key] = &counter{hits: 1, expireAt: now.Add(l.per)}
		return true, 0
	}

	c.hits++
	if c.hits > l.max {
		return false, time.Until(c.expireAt)
	}

	return true, 0
}

// Exceeded reports whether key is already over budget without spending any of
// it. Used where the budget should only be charged for failures.
func (l *Limiter) Exceeded(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	c, ok := l.keys[key]
	if !ok || time.Now().After(c.expireAt) {
		return false, 0
	}

	if c.hits >= l.max {
		return true, time.Until(c.expireAt)
	}

	return false, 0
}

// Forget clears a key's window, so a success can wipe a record of failures.
func (l *Limiter) Forget(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.keys, key)
}

func (l *Limiter) pruneLoop() {
	ticker := time.NewTicker(l.per)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		l.mu.Lock()
		for key, c := range l.keys {
			if now.After(c.expireAt) {
				delete(l.keys, key)
			}
		}
		l.mu.Unlock()
	}
}

func tooManyAttempts(context *gin.Context, retryAfter time.Duration) {
	// Round up, so a sub-second remainder never advertises "retry in 0".
	seconds := int(retryAfter.Seconds()) + 1
	context.Header("Retry-After", strconv.Itoa(seconds))

	context.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"message": "Too many attempts, please try again later",
	})
}

// RateLimitIP caps requests per client IP. Each call builds its own limiter,
// so every route it is attached to gets an independent budget.
//
// This is only as trustworthy as ClientIP(). Gin trusts all proxies by
// default, which lets a caller set X-Forwarded-For and mint a fresh identity
// per request — see the SetTrustedProxies call in main.go.
func RateLimitIP(max int, per time.Duration) gin.HandlerFunc {
	limiter := NewLimiter(max, per)

	return func(context *gin.Context) {
		if ok, retry := limiter.Consume(context.ClientIP()); !ok {
			tooManyAttempts(context, retry)
			return
		}

		context.Next()
	}
}

// ThrottleLoginByAccount limits failed logins per email address, so a single
// account cannot be ground down from a rotating set of IPs. Only failures are
// charged and a success clears the record, so a legitimate user is never
// locked out by their own typos.
func ThrottleLoginByAccount(max int, per time.Duration) gin.HandlerFunc {
	limiter := NewLimiter(max, per)

	return func(context *gin.Context) {
		// ShouldBindBodyWithJSON caches the body in the context, so the
		// handler can still read it after we have peeked at the email.
		var body struct {
			Email string `json:"email"`
		}
		_ = context.ShouldBindBodyWithJSON(&body)

		key := strings.ToLower(strings.TrimSpace(body.Email))
		if key == "" {
			// Malformed body: let the handler reject it. The per-IP budget
			// already covers this case.
			context.Next()
			return
		}

		if exceeded, retry := limiter.Exceeded(key); exceeded {
			tooManyAttempts(context, retry)
			return
		}

		context.Next()

		switch context.Writer.Status() {
		case http.StatusUnauthorized:
			limiter.Consume(key)
		case http.StatusOK:
			limiter.Forget(key)
		}
	}
}
