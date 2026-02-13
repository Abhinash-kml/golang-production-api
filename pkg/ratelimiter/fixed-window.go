package ratelimiter

import (
	"sync"
	"time"
)

type ClientInfo struct {
	Count           int
	WindowStartTime time.Time
}

type FixedWindowLimiter struct {
	WindowDuration time.Duration
	LimitPerWindow int
	Table          map[string]*ClientInfo
	mutex          sync.Mutex
}

func (f *FixedWindowLimiter) Allow(ip string) bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	data, exists := f.Table[ip]
	currentTime := time.Now()

	// If Ip doesnt exist it means its a compleely new request
	if !exists {
		f.Table[ip] = &ClientInfo{Count: 1, WindowStartTime: currentTime}
		return true
	}

	// Window time passed, reset the counter
	if currentTime.Sub(data.WindowStartTime) >= f.WindowDuration {
		data.Count = 1
		data.WindowStartTime = time.Now()
		return true
	}

	// Still within the window, just imcrement the counter
	if data.Count < f.LimitPerWindow {
		data.Count++
		return true
	}

	return false
}

func (f *FixedWindowLimiter) AutoEvict(evictDuration time.Duration) {
	ticker := time.NewTicker(evictDuration)
	for range ticker.C {
		f.mutex.Lock() // This is similar to table level locking - may use row level locking in future
		for key, value := range f.Table {
			if time.Since(value.WindowStartTime) >= evictDuration {
				delete(f.Table, key)
			}
		}
		f.mutex.Unlock()
	}
}
