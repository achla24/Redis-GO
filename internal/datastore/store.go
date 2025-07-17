package datastore

import (
    "sync"
     "fmt"
    "strconv"
    "time"
    // "strings"
)

type Value struct {
    Data      string
    ExpiresAt time.Time // zero means no expiry
}

type Store struct {
    data map[string]Value
    mu   sync.RWMutex
    aof *AOF // Append-Only File for persistence
}

// Constructor
func NewStore(aof *AOF) *Store {
    return &Store{
        data: make(map[string]Value),
        aof: aof,
    }
}

// SET command
func (s *Store) Set(key, value string, ttlSeconds int) {
    s.mu.Lock()
    defer s.mu.Unlock()

    v := Value{
        Data: value,
    }

    if ttlSeconds > 0 {
        v.ExpiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
    }

    s.data[key] = v

    // Write to AOF
    if s.aof != nil {
        ttl := ""
        if ttlSeconds > 0 {
            ttl = " " + strconv.Itoa(ttlSeconds)
        }
        s.aof.WriteCommand(fmt.Sprintf("SET %s %s%s", key, value, ttl))
    }
}

// GET command
func (s *Store) Get(key string) (string, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    val, ok := s.data[key]
    if !ok {
        return "", false
    }

    // Check for TTL expiration
    if !val.ExpiresAt.IsZero() && time.Now().After(val.ExpiresAt) {
        return "", false
    }

    return val.Data, true
}

// DEL command
func (s *Store) Delete(key string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.data, key)

    // Write to AOF
    if s.aof != nil {
        s.aof.WriteCommand(fmt.Sprintf("DEL %s", key))
    }
}

// TTL Cleaner Goroutine
func (s *Store) CleanExpired() {
    s.mu.Lock()
    defer s.mu.Unlock()

    now := time.Now()
    for k, v := range s.data {
        if !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt) {
            delete(s.data, k)
        }
    }
}
