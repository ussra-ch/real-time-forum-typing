package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type LoginRateLimit struct {
	count        int
	FirstTime    time.Time
	BlockedUntil time.Time
	UserIP       string
}

var LoginRateLimits = make(map[string]*LoginRateLimit)

func CheckRateLimitLogin(ratelimit *LoginRateLimit, window time.Duration) bool {
	if time.Now().Before(ratelimit.BlockedUntil) {
		return false
	}
	if time.Now().After(ratelimit.BlockedUntil) && ratelimit.count > 10 {
		ratelimit.FirstTime = time.Now()
		ratelimit.BlockedUntil = time.Time{}
		ratelimit.count = 0
	}
	ratelimit.count++
	if ratelimit.count > 10 {
		ratelimit.BlockedUntil = time.Now().Add(window)
		return false
	}
	return true
}

func UserInfosLogin(r *http.Request) (*LoginRateLimit, bool) {
	rateLimit := &LoginRateLimit{
		count:        0,
		FirstTime:    time.Now(),
		BlockedUntil: time.Time{},
		UserIP:       "",
	}
	userIP := GetUserIP(r)
	rateLimit.UserIP = userIP
	return rateLimit, true
}

func RateLimitLoginMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRateLimit, ok := UserInfosLogin(r)
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

		ratelimit, exists := LoginRateLimits[userRateLimit.UserIP]
		if !exists {
			AddUserToTheMap_Login(userRateLimit)
			ratelimit = userRateLimit
		}

		if !CheckRateLimitLogin(ratelimit, 1*time.Minute) {
			errorr := ErrorStruct{
				Type: "error",
				Text: "Too many requests",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(errorr)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func GetUserIP(r *http.Request) string {
	temp := r.RemoteAddr
	userIP := strings.Split(temp, ":")[0]
	return userIP
}

func AddUserToTheMap_Login(ratelimit *LoginRateLimit) {
	LoginRateLimits[ratelimit.UserIP] = ratelimit
}
