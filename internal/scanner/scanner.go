package scanner

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Host        string
	Ports       []int
	Timeout     time.Duration
	Concurrency int
}

type Result struct {
	Port    int           `json:"port"`
	Open    bool          `json:"open"`
	Latency time.Duration `json:"latency"`
	Banner  string        `json:"banner,omitempty"`
	Detail  string        `json:"detail"`
}

type Summary struct {
	Total  int `json:"total"`
	Open   int `json:"open"`
	Closed int `json:"closed"`
}

func ParsePorts(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("daftar port tidak boleh kosong")
	}

	seen := make(map[int]struct{})
	var ports []int

	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, fmt.Errorf("range port tidak valid: %s", part)
			}

			start, err := parsePort(bounds[0])
			if err != nil {
				return nil, err
			}
			end, err := parsePort(bounds[1])
			if err != nil {
				return nil, err
			}
			if start > end {
				return nil, fmt.Errorf("range port tidak valid: %s", part)
			}

			for port := start; port <= end; port++ {
				if _, ok := seen[port]; ok {
					continue
				}
				seen[port] = struct{}{}
				ports = append(ports, port)
			}
			continue
		}

		port, err := parsePort(part)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[port]; ok {
			continue
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}

	slices.Sort(ports)
	return ports, nil
}

func Run(ctx context.Context, cfg Config) []Result {
	if cfg.Concurrency < 1 {
		cfg.Concurrency = 1
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 800 * time.Millisecond
	}

	results := make([]Result, len(cfg.Ports))
	type job struct {
		index int
		port  int
	}

	jobs := make(chan job)
	var wg sync.WaitGroup

	for range make([]struct{}, cfg.Concurrency) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results[job.index] = scanPort(ctx, cfg.Host, job.port, cfg.Timeout)
			}
		}()
	}

	for index, port := range cfg.Ports {
		jobs <- job{index: index, port: port}
	}
	close(jobs)
	wg.Wait()

	slices.SortFunc(results, func(left, right Result) int {
		if left.Open != right.Open {
			if left.Open {
				return -1
			}
			return 1
		}
		if left.Port < right.Port {
			return -1
		}
		if left.Port > right.Port {
			return 1
		}
		return 0
	})

	return results
}

func FilterOpen(results []Result) []Result {
	filtered := make([]Result, 0, len(results))
	for _, result := range results {
		if result.Open {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func BuildSummary(results []Result) Summary {
	summary := Summary{Total: len(results)}
	for _, result := range results {
		if result.Open {
			summary.Open++
		} else {
			summary.Closed++
		}
	}
	return summary
}

func scanPort(ctx context.Context, host string, port int, timeout time.Duration) Result {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	start := time.Now()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return Result{
			Port:    port,
			Open:    false,
			Latency: time.Since(start),
			Detail:  simplifyError(err),
		}
	}
	defer conn.Close()

	banner := grabBanner(conn, port, timeout)
	detail := "connection established"
	if banner != "" {
		detail = "banner received"
	}

	return Result{
		Port:    port,
		Open:    true,
		Latency: time.Since(start),
		Banner:  banner,
		Detail:  detail,
	}
}

func parsePort(raw string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("port tidak valid: %s", raw)
	}
	if value < 1 || value > 65535 {
		return 0, fmt.Errorf("port di luar rentang: %d", value)
	}
	return value, nil
}

func simplifyError(err error) string {
	text := err.Error()
	if strings.Contains(text, "connection refused") {
		return "connection refused"
	}
	if strings.Contains(text, "i/o timeout") {
		return "timeout"
	}
	return text
}

func grabBanner(conn net.Conn, port int, timeout time.Duration) string {
	_ = conn.SetDeadline(time.Now().Add(timeout))

	switch port {
	case 80, 8080, 8000, 8008, 8888:
		_, _ = conn.Write([]byte("HEAD / HTTP/1.0\r\nHost: localhost\r\n\r\n"))
	case 25, 110, 143:
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	line = strings.TrimSpace(line)
	if len(line) > 120 {
		line = line[:120]
	}
	return line
}
