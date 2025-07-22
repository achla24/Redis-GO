package protocol

import (
	"bufio"
	"strings"
	"testing"
)

func TestPlaintextCommand(t *testing.T) {
	input := "SET mykey myvalue\n"
	reader := bufio.NewReader(strings.NewReader(input))

	cmd, err := ParseCommand(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expected := []string{"SET", "mykey", "myvalue"}
	for i, v := range expected {
		if cmd[i] != v {
			t.Errorf("Expected %v, got %v", v, cmd[i])
		}
	}
}

func TestRESPCommand(t *testing.T) {
	input := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	cmd, err := ParseCommand(reader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expected := []string{"SET", "key", "value"}
	for i, v := range expected {
		if cmd[i] != v {
			t.Errorf("Expected %v, got %v", v, cmd[i])
		}
	}
}

func TestEmptyPlaintextCommand(t *testing.T) {
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	_, err := ParseCommand(reader)
	if err == nil {
		t.Fatal("Expected error for empty command, got none")
	}
}

func TestMalformedRESP(t *testing.T) {
	input := "SET key value\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	peek, _ := reader.Peek(1)
	if peek[0] == '*' {
		t.Fatal("Malformed RESP test: got RESP input, expected plain")
	}

	// Force malformed RESP input
	input = "*2\r\n$3\r\nSET\r\n$5\r\n" // Incomplete RESP
	reader = bufio.NewReader(strings.NewReader(input))

	_, err := ParseRESP(reader)
	if err == nil {
		t.Fatal("Expected error for malformed RESP input, got none")
	}
}
