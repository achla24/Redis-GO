//entry point to run the server
package main

import (
	"fmt"
	"time"
	"redis-go/internal/datastore"

	"log"
	"net"
	"redis-go/internal/server"
)

func main() {
    store := datastore.NewStore()

    // Start TTL cleaner in background
    go func() {
        for {
            time.Sleep(1 * time.Second)
            store.CleanExpired()
        }
    }()

	//start TCP server
	ln,err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	fmt.Println("Server started on port 6379")
    // fmt.Println("Running local key-value store...")

	for { //infinite loop => server run forever
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		log.Println("New client connected:", conn.RemoteAddr())
		go server.HandleConnection(conn, store) //new goroutine for new client
	}


    // Test SET with TTL
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