package network

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/teiyou416/simblock_go/core"
)

// LoadLatencyMatrix loads a whitespace-separated latency matrix from file.
func LoadLatencyMatrix(path string) ([][]core.SimTime, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open latency matrix: %w", err)
	}
	defer f.Close()

	return ParseLatencyMatrix(f)
}

// ParseLatencyMatrix parses a matrix where each non-empty line is one row.
//
// Rules:
// - comments start with '#'
// - values are integers >= 0
// - matrix must be square (N x N)
func ParseLatencyMatrix(r io.Reader) ([][]core.SimTime, error) {
	scanner := bufio.NewScanner(r)
	var matrix [][]core.SimTime

	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		row := make([]core.SimTime, 0, len(fields))
		for _, field := range fields {
			v, err := strconv.ParseInt(field, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid latency value %q: %w", field, err)
			}
			if v < 0 {
				return nil, fmt.Errorf("latency must be >= 0, got %d", v)
			}
			row = append(row, core.SimTime(v))
		}
		matrix = append(matrix, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan latency matrix: %w", err)
	}
	if len(matrix) == 0 {
		return nil, fmt.Errorf("latency matrix is empty")
	}

	expected := len(matrix)
	for i, row := range matrix {
		if len(row) != expected {
			return nil, fmt.Errorf("row %d has %d columns, want %d", i, len(row), expected)
		}
	}

	return matrix, nil
}
