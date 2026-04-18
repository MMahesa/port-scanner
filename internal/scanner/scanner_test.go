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

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("SSH-2.0-test-server\r\n"))
	}()

	openPort := listener.Addr().(*net.TCPAddr).Port
	closedPort := openPort + 1
	progressCalls := 0

	results := Run(context.Background(), Config{
		Host:        "127.0.0.1",
		Ports:       []int{openPort, closedPort},
		Timeout:     300 * time.Millisecond,
		Concurrency: 2,
		OnProgress: func(done, total int) {
			progressCalls++
			if total != 2 {
				t.Fatalf("expected total 2, got %d", total)
			}
		},
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if progressCalls != 2 {
		t.Fatalf("expected progress callback twice, got %d", progressCalls)
	}
	if !results[0].Open {
		t.Fatalf("expected port %d to be open", openPort)
	}
	if !strings.Contains(results[0].Banner, "SSH-2.0-test-server") {
		t.Fatalf("expected banner on open port, got %q", results[0].Banner)
	}
	if results[1].Open {
		t.Fatalf("expected port %d to be closed", closedPort)
	}
	if !strings.Contains(results[1].Detail, "connection refused") && results[1].Detail == "" {
		t.Fatalf("expected detail for closed port, got %q", results[1].Detail)
	}
}

func TestResultsToCSV(t *testing.T) {
	data, err := ResultsToCSV([]Result{
		{Port: 22, Open: true, Latency: 12 * time.Millisecond, Detail: "banner received", Banner: "SSH-2.0-test"},
	})
	if err != nil {
		t.Fatal(err)
	}

	text := string(data)
	if !strings.Contains(text, "port,status,latency,detail,banner") {
		t.Fatalf("expected csv header, got %q", text)
	}
	if !strings.Contains(text, "22,open,12ms,banner received,SSH-2.0-test") {
		t.Fatalf("expected csv row, got %q", text)
	}
}
