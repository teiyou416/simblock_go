package network

import (
	"strings"
	"testing"

	"github.com/teiyou416/simblock_go/core"
)

func TestParseLatencyMatrix(t *testing.T) {
	const in = `
	# example N x N matrix
	32 124 184
	124 11 227
	184 227 88
	`

	got, err := ParseLatencyMatrix(strings.NewReader(in))
	if err != nil {
		t.Fatalf("ParseLatencyMatrix() err=%v", err)
	}

	want := [][]core.SimTime{
		{32, 124, 184},
		{124, 11, 227},
		{184, 227, 88},
	}
	for i := range want {
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("matrix[%d][%d]: got=%d want=%d", i, j, got[i][j], want[i][j])
			}
		}
	}
}

func TestParseLatencyMatrixRejectsNonSquare(t *testing.T) {
	const in = `
	1 2
	3 4 5
	`
	_, err := ParseLatencyMatrix(strings.NewReader(in))
	if err == nil {
		t.Fatal("ParseLatencyMatrix() err=nil, want non-square error")
	}
}
