package network

import "testing"

func TestDefaultRegions(t *testing.T) {
	if got, want := len(DefaultRegions), 6; got != want {
		t.Fatalf("region count: got=%d want=%d", got, want)
	}

	wantNames := []string{
		"NORTH_AMERICA",
		"EUROPE",
		"SOUTH_AMERICA",
		"ASIA_PACIFIC",
		"JAPAN",
		"AUSTRALIA",
	}
	gotNames := DefaultRegionNames()
	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("region name at %d: got=%s want=%s", i, gotNames[i], wantNames[i])
		}
	}
}
