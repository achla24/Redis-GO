package datastore

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Subscriber struct {
	conn net.Conn
	ch   chan string
}

type Value struct {
	Data      string
	ExpiresAt time.Time // zero means no expiry

}

type Store struct {
	data map[string]Value
	mu   sync.RWMutex
	// aof  *AOF // Append-Only File for persistence
	AOF         *AOF
	subscribers map[string][]*Subscriber
	// subMu       sync.RWMutex
}

func NewStore(aof *AOF) *Store {
	return &Store{
		data: make(map[string]Value),
		// initialize other fields as needed
		// AOF:         NewAOFLogger("appendonly.aof"),

		AOF:         aof, // Use the provided AOF logger passed from main.go
		subscribers: make(map[string][]*Subscriber),
	}
}

// Constructor
// func NewStore(aof *AOF) *Store {
// 	return &Store{
// 		data: make(map[string]Value),
// 		aof:  aof,
// 	}
// }

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
	if s.AOF != nil {
		ttl := ""
		if ttlSeconds > 0 {
			ttl = " " + strconv.Itoa(ttlSeconds)
		}
		s.AOF.WriteCommand(fmt.Sprintf("SET %s %s%s", key, value, ttl))
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
	if s.AOF != nil {
		s.AOF.WriteCommand(fmt.Sprintf("DEL %s", key))
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

// for pub/sub

// var pubSubRegistry = struct {
// 	sync.RWMutex
// 	subscribers map[string][]net.Conn
// }{subscribers: make(map[string][]net.Conn)}

// func Subscribe(channel string, conn net.Conn) {
// 	pubSubRegistry.Lock()
// 	defer pubSubRegistry.Unlock()
// 	pubSubRegistry.subscribers[channel] = append(pubSubRegistry.subscribers[channel], conn)
// }

func (s *Store) Subscribe(channel string, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Prevent duplicate subscriptions
	for _, sub := range s.subscribers[channel] {
		if sub.conn == conn {
			return // Already subscribed
		}
	}

	sub := &Subscriber{
		conn: conn,
		ch:   make(chan string, 10),
	}
	s.subscribers[channel] = append(s.subscribers[channel], sub)

	go func() {
		for msg := range sub.ch {
			fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(msg), msg)
		}
	}()
}

// func Publish(channel string, message string) {
// 	pubSubRegistry.RLock()
// 	defer pubSubRegistry.RUnlock()
// 	for _, conn := range pubSubRegistry.subscribers[channel] {
// 		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(message), message) // RESP format
// 	}
// }

func (s *Store) Publish(channel, message string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sub := range s.subscribers[channel] {
		sub.ch <- message
	}
}

func (s *Store) CloseAOF() {
	if s.AOF != nil {
		s.AOF.Close()
	}
}

func (a *AOFLogger) Close() error {
	if a.file != nil {
		return a.file.Close()
	}
	return nil
}

func ListenForShutdown(store *Store) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		<-c
		fmt.Println("Shutting down...")
		store.CloseAOF()
		os.Exit(0)
	}()
}
