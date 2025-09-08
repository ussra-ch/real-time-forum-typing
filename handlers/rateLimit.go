package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type RateLimit struct {
	count        int
	FirstTime    time.Time
	BlockedUntil time.Time
	UserId       int
}

type ErrorStruct struct {
	Type string
	Text string
}

var (
	CommentRateLimits = make(map[int]*RateLimit)
	PostRateLimits    = make(map[int]*RateLimit)
)

func CheckRateLimit(ratelimit *RateLimit, window time.Duration, maxAttempts int) bool {
	if time.Now().Before(ratelimit.BlockedUntil) {
		return false
	}
	if time.Now().After(ratelimit.BlockedUntil) && ratelimit.count > maxAttempts {
		ratelimit.FirstTime = time.Now()
		ratelimit.BlockedUntil = time.Time{}
		ratelimit.count = 0
	}
	ratelimit.count++
	if ratelimit.count > maxAttempts {
		ratelimit.BlockedUntil = time.Now().Add(window)
		return false
	}
	return true
}

func UserInfos(r *http.Request) (*RateLimit, bool) {
	rateLimit := &RateLimit{
		count:        0,
		FirstTime:    time.Now(),
		BlockedUntil: time.Time{},
		UserId:       -1,
	}
	mu.Lock()
	_, userID := IsLoggedIn(r)
	mu.Unlock()
	rateLimit.UserId = userID
	return rateLimit, true
}

func RatelimitMiddleware(next http.HandlerFunc, rateLimitType string, maxAttempts int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		theMap := make(map[int]*RateLimit)
		if rateLimitType == "posts" {
			theMap = PostRateLimits
		} else if rateLimitType == "comments" {
			theMap = CommentRateLimits
		}
		userRateLimit, ok := UserInfos(r)

		if !ok {
			errorr := ErrorStruct{
				Type: "error",
				Text: "Unauthorized",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(errorr)
			return
		}

		ratelimit, exists := theMap[userRateLimit.UserId]
		if !exists {
			mu.Lock()
			AddUserToTheMap(userRateLimit, theMap)
			mu.Unlock()
			ratelimit = userRateLimit
		}

		if !CheckRateLimit(ratelimit, 1*time.Minute, maxAttempts) {
			errorHandler(http.StatusTooManyRequests, w)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func AddUserToTheMap(ratelimit *RateLimit, theMap map[int]*RateLimit) {
	theMap[ratelimit.UserId] = ratelimit
}
