package server

import (
	"bufio"
	"net"
	"redis-go/internal/datastore"
	"redis-go/internal/protocol"
	"strconv"
	"strings"
)

func HandleConnection(conn net.Conn, store *datastore.Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		
		

		parts, err := protocol.ParseCommand(reader)
		if err != nil {
			conn.Write([]byte("-ERR connection closed\r\n"))
			return
		}

		
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])

		switch cmd {

		case "PING":
			conn.Write([]byte(protocol.SerializeSimpleString("PONG")))

		case "SET":
			ttl := 0

			if len(parts) == 5 && strings.ToUpper(parts[3]) == "EX" {
				parsedTTL, err := strconv.Atoi(parts[4])
				if err != nil {
					conn.Write([]byte(protocol.SerializeError("ERR invalid TTL")))
					continue
				}
				ttl = parsedTTL
			} else if len(parts) != 3 {
				conn.Write([]byte(protocol.SerializeError("ERR wrong number of arguments for 'SET' command")))
				continue
			}

			store.Set(parts[1], parts[2], ttl)
			conn.Write([]byte(protocol.SerializeSimpleString("OK")))

		case "GET":
			if len(parts) != 2 {
				conn.Write([]byte(protocol.SerializeError("ERR wrong number of arguments for 'GET' command")))
				continue
			}
			val, ok := store.Get(parts[1])
			if !ok {
				conn.Write([]byte(protocol.SerializeBulkString(""))) // nil response
			} else {
				conn.Write([]byte(protocol.SerializeBulkString(val)))
			}

		case "SUBSCRIBE":
			// Only works if you have a Subscribe function in datastore
			if len(parts) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for SUBSCRIBE\r\n"))
				continue
			}
			// store.Subscribe(parts[1], conn)
			// conn.Write([]byte("+Subscribed to " + parts[1] + "\r\n"))
			channel := parts[1]
			conn.Write([]byte("+Subscribed to " + channel + "\r\n"))
			store.Subscribe(channel, conn)

			// Keep connection alive so it can receive messages
			select {}

		case "PUBLISH":
			if len(parts) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for PUBLISH\r\n"))
				continue
			}
			channel := parts[1]
			// message := parts[2]
			// store.Publish(channel, message)

			// // Log to AOF
			// if datastore.ShouldLogToAOF("PUBLISH") && store != nil {
			//     store.AOF.WriteCommand("PUBLISH", channel, message)
			// }

			// conn.Write([]byte(":1\r\n"))
			message := strings.Join(parts[2:], " ")
			store.Publish(channel, message)
			conn.Write([]byte("+Message published to " + channel + "\r\n"))

		default:
			conn.Write([]byte(protocol.SerializeError("ERR unknown command")))
		}
	}
}

// package server

// import (
// 	"bufio"
// 	// "fmt"
// 	"net"
// 	"redis-go/internal/datastore"
// 	"redis-go/internal/protocol"
// 	"strconv"
// 	"strings"
// )

// func HandleConnection(conn net.Conn, store *datastore.Store) {
// 	defer conn.Close()
// 	reader := bufio.NewReader(conn)

// 	for {
// 		line, err := reader.ReadString('\n')
// 		if err != nil {
// 			conn.Write([]byte("-ERR connection closed\r\n"))
// 			return
// 		}

// 		line = strings.TrimSpace(line)
// 		if line == "" {
// 			continue
// 		}

// 		parts := strings.Fields(line)
// 		cmd := strings.ToUpper(parts[0])

// 		switch cmd {

// 		case "PING":
// 			conn.Write([]byte(protocol.SerializeSimpleString("PONG")))

// 		case "SET":
// 			ttl := 0

// 			if len(parts) == 5 && strings.ToUpper(parts[3]) == "EX" {
// 				parsedTTL, err := strconv.Atoi(parts[4])
// 				if err != nil {
// 					conn.Write([]byte(protocol.SerializeError("ERR invalid TTL")))
// 					continue
// 				}
// 				ttl = parsedTTL
// 			} else if len(parts) != 3 {
// 				conn.Write([]byte(protocol.SerializeError("ERR wrong number of arguments for 'SET' command")))
// 				continue
// 			}

// 			store.Set(parts[1], parts[2], ttl)

// 			conn.Write([]byte(protocol.SerializeSimpleString("OK")))

// 		case "GET":
// 			if len(parts) != 2 {
// 				conn.Write([]byte(protocol.SerializeError("ERR wrong number of arguments for 'GET' command")))
// 				continue
// 			}
// 			val, ok := store.Get(parts[1])
// 			if !ok {
// 				conn.Write([]byte(protocol.SerializeBulkString(""))) // nil response
// 			} else {
// 				conn.Write([]byte(protocol.SerializeBulkString(val)))
// 			}

// 		case "SUBSCRIBE":
// 			if len(parts) != 2 {
// 				conn.Write([]byte("-ERR wrong number of arguments for SUBSCRIBE\r\n"))
// 				continue
// 			}
// 			datastore.Subscribe(parts[1], conn)
// 			conn.Write([]byte("+Subscribed to " + parts[1] + "\r\n"))
// 		case "PUBLISH":
// 			if len(parts) != 3 {
// 				conn.Write([]byte("-ERR wrong number of arguments for PUBLISH\r\n"))
// 				continue
// 			}
// 			datastore.Publish(parts[1], parts[2])
// 			conn.Write([]byte(":1\r\n"))

// 		default:
// 			conn.Write([]byte(protocol.SerializeError("ERR unknown command")))
// 		}
// 	}
// }

// //---------------------------------//

// // func HandleConnection(conn net.Conn, store *datastore.Store) {
// // 	defer conn.Close()
// // 	reader := bufio.NewReader(conn)

// // 	for {
// // 		line, err := reader.ReadString('\n')
// // 		if err != nil {
// // 			fmt.Fprintln(conn, "ERR: connection closed")
// // 			return
// // 		}

// // 		line = strings.TrimSpace(line)
// // 		if line == "" {
// // 			continue
// // 		}

// // 		parts := strings.Fields(line)
// // 		cmd := strings.ToUpper(parts[0])

// // 		switch cmd {
// // 		case "SET":
// // 			if len(parts) < 3 {
// // 				fmt.Fprintln(conn, "ERR: usage SET key value")
// // 				continue
// // 			}
// // 			key, value := parts[1], parts[2]
// // 			store.Set(key, value, 0)
// // 			fmt.Fprintln(conn, "OK")

// // 		case "GET":
// // 			if len(parts) < 2 {
// // 				fmt.Fprintln(conn, "ERR: usage GET key")
// // 				continue
// // 			}
// // 			key := parts[1]
// // 			if val, ok := store.Get(key); ok {
// // 				fmt.Fprintln(conn, val)
// // 			} else {
// // 				fmt.Fprintln(conn, "(nil)")
// // 			}

// // 		case "DEL":
// // 			if len(parts) < 2 {
// // 				fmt.Fprintln(conn, "ERR: usage DEL key")
// // 				continue
// // 			}
// // 			key := parts[1]
// // 			store.Delete(key)
// // 			fmt.Fprintln(conn, "OK")

// // 		default:
// // 			fmt.Fprintln(conn, "ERR: unknown command")
// // 		}
// // 	}
// // }
