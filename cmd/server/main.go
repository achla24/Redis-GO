//entry point to run the server
package main

import (
	"fmt"
	"time"
	"redis-go/internal/datastore"
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

    fmt.Println("Running local key-value store...")

    // Test SET with TTL
    store.Set("name", "Achla", 5) // TTL 5 seconds
    fmt.Println("SET name = Achla (expires in 5s)")

    // Test GET
    val, ok := store.Get("name")
    if ok {
        fmt.Println("GET name:", val)
    } else {
        fmt.Println("GET name: not found")
    }

    time.Sleep(6 * time.Second)

    // Test expired key
    val, ok = store.Get("name")
    if ok {
        fmt.Println("GET name after TTL:", val)
    } else {
        fmt.Println("GET name after TTL: not found (expired)")
    }
}