package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucketClientInfo struct {
	tokens      int
	lastChecked time.Time
}

type TokenBucketLimiter struct {
	Table         map[string]*TokenBucketClientInfo
	Capacity      int
	InitialTokens int
	RefillRate    time.Duration
	mutex         sync.Mutex
}

func (f *TokenBucketLimiter) Allow(ip string) bool {
	data, exists := f.Table[ip]
	if !exists {
		f.Table[ip] = &TokenBucketClientInfo{tokens: f.InitialTokens, lastChecked: time.Now()}
		return true
	}

	elapsedTime := time.Since(data.lastChecked)
	tokensToAdd := int(elapsedTime / f.RefillRate)
	data.tokens = min(f.Capacity, data.tokens+tokensToAdd)
	data.lastChecked = time.Now()

	if data.tokens > 0 {
		data.tokens--
		return true
	}

	return false
}
