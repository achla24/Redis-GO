//Parses RESP commands from the client using bufio.Reader

package protocol

import (
	"bufio"
	
	"fmt"
	"strconv"
	"strings"
)

// ParseRESP parses a RESP command from the connection
func ParseRESP(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("expected array")
	}

	numArgs, _ := strconv.Atoi(strings.TrimSpace(line[1:]))

	args := make([]string, 0, numArgs)
	for i := 0; i < numArgs; i++ {
		// Read bulk string length line
		_, err := reader.ReadString('\n') // Skip $length line
		if err != nil {
			return nil, err
		}

		// Read actual data line
		arg, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		args = append(args, strings.TrimSpace(arg))
	}
	return args, nil
}
