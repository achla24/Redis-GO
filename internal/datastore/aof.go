package datastore

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type AOFLogger struct {
	file *os.File
}

func NewAOFLogger(filename string) *AOFLogger {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open AOF file: %v", err))
	}
	return &AOFLogger{file: f}
}

func (a *AOFLogger) WriteCommand(cmd string, args ...string) error {
	line := cmd
	for _, arg := range args {
		line += " " + arg
	}
	line += "\n"
	_, err := a.file.WriteString(line)
	return err
}

var aofIgnoredCommands = map[string]bool{
	"SUBSCRIBE": true,
	// "PUBLISH":   true,
}

func ShouldLogToAOF(command string) bool {
	_, ok := aofIgnoredCommands[strings.ToUpper(command)]
	return !ok
}

type AOF struct {
	file *os.File
	mu   sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AOF{file: f}, nil
}

func LoadAOF(store *Store, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])
		switch cmd {
		case "SET":
			if len(parts) < 3 {
				return fmt.Errorf("invalid SET command: %v", line)
			}
			key := parts[1]
			value := parts[2]
			ttl := int64(0)
			if len(parts) == 4 {
				parsedTTL, err := strconv.ParseInt(parts[3], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid TTL in AOF: %v", line)
				}
				ttl = parsedTTL
			}
			store.Set(key, value, int(ttl))

		case "DEL":
			if len(parts) < 2 {
				return fmt.Errorf("invalid DEL command: %v", line)
			}
			store.Delete(parts[1])

		default:
			// return fmt.Errorf("unsupported command in AOF: %s", cmd)

			//fix(aof): skip unsupported PUBLISH commands to prevent server crash on startup
			log.Printf("Skipping unsupported command in AOF: %s", line)
			continue
		}
	}

	return scanner.Err()
}

// WriteCommand formats and writes a command to AOF: e.g., SET key value 60
func (a *AOF) WriteCommand(cmd string, args ...interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	strArgs := make([]string, len(args))
	for i, arg := range args {
		strArgs[i] = fmt.Sprintf("%v", arg)
	}

	line := cmd + " " + strings.Join(strArgs, " ") + "\n"
	_, err := a.file.WriteString(line)
	return err
}

func (a *AOF) Close() error {
	return a.file.Close()
}
