// entry point to run the server
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"redis-go/internal/datastore"
	"redis-go/internal/server"
	"syscall"
	"time"
)

func main() {
	// 1. Initialize AOF Logger
	aofLogger, err := datastore.NewAOF("../appendonly.aof") //added ../ coz otherwise aof text file will be added in cmd/server
	if err != nil {
		fmt.Printf("Failed to open AOF file: %v", err)
		return
	}
	defer aofLogger.Close()

	// 2. Initialize the store with AOF logger
	store := datastore.NewStore(aofLogger)

	// 3. Load commands from AOF at startup
	if err := datastore.LoadAOF(store, "../appendonly.aof"); err != nil {
		log.Fatalf("Failed to load AOF commands: %v", err)
	}

	// store := datastore.NewStore()

	// Start TTL cleaner in background
	go func() {
		for {
			time.Sleep(1 * time.Second)
			store.CleanExpired()
		}
	}()

	//start TCP server
	fmt.Println("main.go starting")
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	fmt.Println("Server started on port 6379")
	os.Stdout.Sync()
	// fmt.Println("Running local key-value store...")

	//for smooth shutdown
	go handleShutdown()

	for { //infinite loop => server run forever
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		log.Println("New client connected:", conn.RemoteAddr())
		go server.HandleConnection(conn, store) //new goroutine for new client connection
	}

	// // Test SET with TTL
	// store.Set("name", "Achla", 5) // TTL 5 seconds
	// fmt.Println("SET name = Achla (expires in 5s)")

	// // Test GET
	// val, ok := store.Get("name")
	// if ok {
	//     fmt.Println("GET name:", val)
	// } else {
	//     fmt.Println("GET name: not found")
	// }

	// time.Sleep(6 * time.Second)

	// // Test expired key
	// val, ok = store.Get("name")
	// if ok {
	//     fmt.Println("GET name after TTL:", val)
	// } else {
	//     fmt.Println("GET name after TTL: not found (expired)")
	// }
}
func handleShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Gracefully shutting down...")
	// FlushAOF() // if needed
	os.Exit(0)
}
