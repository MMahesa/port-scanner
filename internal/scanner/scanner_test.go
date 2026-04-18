package scanner

import (
	"context"
	"net"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParsePorts(t *testing.T) {
	ports, err := ParsePorts("22,80,443,8000-8002,80")
	if err != nil {
		t.Fatal(err)
	}

	expected := []int{22, 80, 443, 8000, 8001, 8002}
	if !slices.Equal(ports, expected) {
		t.Fatalf("expected %v, got %v", expected, ports)
	}
}

func TestParsePortsRejectsInvalidRange(t *testing.T) {
	_, err := ParsePorts("100-90")
	if err == nil {
		t.Fatal("expected error for invalid range")
	}
}

func TestRunScansPorts(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	openPort := listener.Addr().(*net.TCPAddr).Port
	closedPort := openPort + 1

	results := Run(context.Background(), Config{
		Host:        "127.0.0.1",
		Ports:       []int{openPort, closedPort},
		Timeout:     300 * time.Millisecond,
		Concurrency: 2,
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Open {
		t.Fatalf("expected port %d to be open", openPort)
	}
	if results[1].Open {
		t.Fatalf("expected port %d to be closed", closedPort)
	}
	if !strings.Contains(results[1].Detail, "connection refused") && results[1].Detail == "" {
		t.Fatalf("expected detail for closed port, got %q", results[1].Detail)
	}
}
