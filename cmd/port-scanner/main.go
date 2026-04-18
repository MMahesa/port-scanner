package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MMahesa/port-scanner/internal/scanner"
)

func main() {
	host := flag.String("host", "", "host atau IP target")
	ports := flag.String("ports", "22,80,443", "daftar port, contoh: 22,80,443,8000-8100")
	timeout := flag.Duration("timeout", 800*time.Millisecond, "timeout koneksi per port")
	concurrency := flag.Int("concurrency", 200, "jumlah worker untuk scan paralel")
	format := flag.String("format", "table", "format output: table, json, atau csv")
	output := flag.String("output", "", "simpan hasil ke file")
	onlyOpen := flag.Bool("open-only", false, "tampilkan hanya port yang terbuka")
	flag.Parse()

	if strings.TrimSpace(*host) == "" {
		fmt.Fprintln(os.Stderr, "host wajib diisi, contoh: --host scanme.nmap.org")
		os.Exit(1)
	}

	portList, err := scanner.ParsePorts(*ports)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg := scanner.Config{
		Host:        strings.TrimSpace(*host),
		Ports:       portList,
		Timeout:     *timeout,
		Concurrency: *concurrency,
	}
	if strings.EqualFold(strings.TrimSpace(*format), "table") && strings.TrimSpace(*output) == "" {
		cfg.OnProgress = func(done, total int) {
			fmt.Fprintf(os.Stderr, "\rMemindai port: %d/%d", done, total)
			if done == total {
				fmt.Fprintln(os.Stderr)
			}
		}
	}

	results := scanner.Run(context.Background(), cfg)
	if *onlyOpen {
		results = scanner.FilterOpen(results)
	}

	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "json":
		payload := map[string]any{
			"target":  cfg.Host,
			"summary": scanner.BuildSummary(results),
			"results": results,
		}
		writer, closeFn, err := outputWriter(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer closeFn()

		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(payload); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "csv":
		data, err := scanner.ResultsToCSV(results)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *output == "" {
			_, err = os.Stdout.Write(data)
		} else {
			err = os.WriteFile(*output, data, 0o644)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		text := renderTable(cfg.Host, results)
		if *output == "" {
			fmt.Print(text)
			return
		}
		if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func renderTable(host string, results []scanner.Result) string {
	summary := scanner.BuildSummary(results)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Target: %s\n", host)
	fmt.Fprintf(&builder, "%-8s %-8s %-12s %-16s %s\n", "PORT", "STATUS", "LATENCY", "DETAIL", "BANNER")
	for _, result := range results {
		status := "closed"
		if result.Open {
			status = "open"
		}
		fmt.Fprintf(&builder, "%-8d %-8s %-12s %-16s %s\n",
			result.Port,
			status,
			result.Latency.Round(time.Millisecond),
			result.Detail,
			result.Banner,
		)
	}
	fmt.Fprintln(&builder)
	fmt.Fprintf(&builder, "Ringkasan: total=%d terbuka=%d tertutup=%d\n", summary.Total, summary.Open, summary.Closed)
	return builder.String()
}

func outputWriter(path string) (*os.File, func(), error) {
	if strings.TrimSpace(path) == "" {
		return os.Stdout, func() {}, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return file, func() { _ = file.Close() }, nil
}
