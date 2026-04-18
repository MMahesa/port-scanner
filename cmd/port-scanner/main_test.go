package main

import (
	"strings"
	"testing"
	"time"

	"github.com/MMahesa/port-scanner/internal/scanner"
)

func TestRenderTable(t *testing.T) {
	output := renderTable("127.0.0.1", []scanner.Result{
		{Port: 22, Open: true, Latency: 12 * time.Millisecond, Detail: "connection established"},
		{Port: 80, Open: false, Latency: 33 * time.Millisecond, Detail: "connection refused"},
	})

	if !strings.Contains(output, "Target: 127.0.0.1") {
		t.Fatalf("expected host line, got %q", output)
	}
	if !strings.Contains(output, "Ringkasan: total=2 terbuka=1 tertutup=1") {
		t.Fatalf("expected summary line, got %q", output)
	}
}
