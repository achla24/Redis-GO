//Converts Go responses into RESP format for the client.

package protocol

import (
	"fmt"
)

// Simple string: +OK\r\n
func SerializeSimpleString(message string) string {
	return fmt.Sprintf("+%s\r\n", message)
}

// Bulk string: $5\r\nvalue\r\n
func SerializeBulkString(message string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(message), message)
}

// Error: -Error message\r\n
func SerializeError(message string) string {
	return fmt.Sprintf("-%s\r\n", message)
}
