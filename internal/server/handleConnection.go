package server

import (
	"bufio"
	"fmt"
	"net"
	"redis-go/internal/datastore"
	"strings"
)

func HandleConnection(conn net.Conn, store *datastore.Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			conn.Write([]byte("-ERR connection closed\r\n"))
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "SET":
			if len(parts) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SET'\r\n"))
				continue
			}
			key, value := parts[1], parts[2]
			store.Set(key, value, 0)
			conn.Write([]byte("+OK\r\n"))

		case "GET":
			if len(parts) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'GET'\r\n"))
				continue
			}
			key := parts[1]
			if val, ok := store.Get(key); ok {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))) // RESP bulk string
			} else {
				conn.Write([]byte("$-1\r\n")) // RESP nil
			}

		case "DEL":
			if len(parts) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'DEL'\r\n"))
				continue
			}
			key := parts[1]
			store.Delete(key)
			conn.Write([]byte("+OK\r\n"))

		default:
			conn.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}

// func HandleConnection(conn net.Conn, store *datastore.Store) {
// 	defer conn.Close()
// 	reader := bufio.NewReader(conn)

// 	for {
// 		line, err := reader.ReadString('\n')
// 		if err != nil {
// 			fmt.Fprintln(conn, "ERR: connection closed")
// 			return
// 		}

// 		line = strings.TrimSpace(line)
// 		if line == "" {
// 			continue
// 		}

// 		parts := strings.Fields(line)
// 		cmd := strings.ToUpper(parts[0])

// 		switch cmd {
// 		case "SET":
// 			if len(parts) < 3 {
// 				fmt.Fprintln(conn, "ERR: usage SET key value")
// 				continue
// 			}
// 			key, value := parts[1], parts[2]
// 			store.Set(key, value, 0)
// 			fmt.Fprintln(conn, "OK")

// 		case "GET":
// 			if len(parts) < 2 {
// 				fmt.Fprintln(conn, "ERR: usage GET key")
// 				continue
// 			}
// 			key := parts[1]
// 			if val, ok := store.Get(key); ok {
// 				fmt.Fprintln(conn, val)
// 			} else {
// 				fmt.Fprintln(conn, "(nil)")
// 			}

// 		case "DEL":
// 			if len(parts) < 2 {
// 				fmt.Fprintln(conn, "ERR: usage DEL key")
// 				continue
// 			}
// 			key := parts[1]
// 			store.Delete(key)
// 			fmt.Fprintln(conn, "OK")

// 		default:
// 			fmt.Fprintln(conn, "ERR: unknown command")
// 		}
// 	}
// }
