package server

import (
	"bufio"
	"net"
	"redis-go/internal/datastore"
	"redis-go/internal/protocol"
	"strconv" //for ascii to int conversion of ttl value in set
	"strings"
)

func HandleConnection(conn net.Conn, store *datastore.Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := protocol.ParseRESP(reader)
		if err != nil {
			conn.Write([]byte(protocol.SerializeError("invalid command")))
			return
		}

		if len(args) == 0 {
			continue
		}

		cmd := strings.ToUpper(args[0])

		switch cmd {
		case "PING":
			conn.Write([]byte(protocol.SerializeSimpleString("PONG")))
		case "SET":
			ttl := 0

			if len(args) == 5 && strings.ToUpper(args[3]) == "EX" {
				parsedTTL, err := strconv.Atoi(args[4])
				if err != nil {
					conn.Write([]byte(protocol.SerializeError("ERR invalid TTL")))
					continue
				}
				ttl = parsedTTL
			} else if len(args) != 3 {
				conn.Write([]byte(protocol.SerializeError("ERR wrong number of arguments for 'SET' command")))
				continue
			}

			store.Set(args[1], args[2], ttl)

			conn.Write([]byte(protocol.SerializeSimpleString("OK")))
		case "GET":
			if len(args) != 2 {
				conn.Write([]byte(protocol.SerializeError("ERR wrong number of arguments for 'GET' command")))
				continue
			}
			val, ok := store.Get(args[1])
			if !ok {
				conn.Write([]byte(protocol.SerializeBulkString(""))) // nil response
			} else {
				conn.Write([]byte(protocol.SerializeBulkString(val)))
			}
		default:
			conn.Write([]byte(protocol.SerializeError("ERR unknown command")))
		}
	}
}
