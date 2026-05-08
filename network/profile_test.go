package network

import "testing"

func TestProfileByNameDefaultsToBitcoin2019(t *testing.T) {
	profile, ok := ProfileByName("")
	if !ok {
		t.Fatal("ProfileByName(\"\") returned ok=false")
	}
	if profile.Name != "bitcoin_2019" {
		t.Fatalf("profile name: got=%q want=%q", profile.Name, "bitcoin_2019")
	}
}

func TestBitcoin2019ProfileBandwidths(t *testing.T) {
	upload, download := Bitcoin2019Profile.Bandwidths(7)

	if got, want := len(upload), 7; got != want {
		t.Fatalf("upload len: got=%d want=%d", got, want)
	}
	if got, want := len(download), 7; got != want {
		t.Fatalf("download len: got=%d want=%d", got, want)
	}
	if got, want := upload[0], uint64(19_200_000); got != want {
		t.Fatalf("upload[0]: got=%d want=%d", got, want)
	}
	if got, want := download[0], uint64(52_000_000); got != want {
		t.Fatalf("download[0]: got=%d want=%d", got, want)
	}
	if got, want := upload[6], upload[0]; got != want {
		t.Fatalf("upload repeats: got=%d want=%d", got, want)
	}
}

func TestProfileWithOverrides(t *testing.T) {
	profile := Bitcoin2019Profile.WithOverrides(ProfileOverrides{
		UploadBandwidth:    []uint64{1, 2},
		DownloadBandwidth:  []uint64{3, 4},
		RegionDistribution: []float64{0.25, 0.75},
		DegreeDistribution: []float64{0.5, 1.0},
	})

	if got, want := profile.UploadBandwidth[0], uint64(1); got != want {
		t.Fatalf("upload override: got=%d want=%d", got, want)
	}
	if got, want := profile.DownloadBandwidth[1], uint64(4); got != want {
		t.Fatalf("download override: got=%d want=%d", got, want)
	}
	if got, want := profile.RegionDistribution[1], 0.75; got != want {
		t.Fatalf("region distribution override: got=%f want=%f", got, want)
	}
	if got, want := profile.DegreeDistribution[0], 0.5; got != want {
		t.Fatalf("degree distribution override: got=%f want=%f", got, want)
	}
}

func TestProfileValidate(t *testing.T) {
	profile := Bitcoin2019Profile
	if err := profile.Validate(6); err != nil {
		t.Fatalf("Validate(6) err=%v", err)
	}

	badBandwidth := profile
	badBandwidth.UploadBandwidth = []uint64{1}
	if err := badBandwidth.Validate(6); err == nil {
		t.Fatal("expected bandwidth length validation error")
	}

	badRegionDistribution := profile
	badRegionDistribution.UploadBandwidth = []uint64{1, 2, 3}
	badRegionDistribution.DownloadBandwidth = []uint64{1, 2, 3}
	badRegionDistribution.RegionDistribution = []float64{0.5, 0.4, 0.2}
	if err := badRegionDistribution.Validate(3); err == nil {
		t.Fatal("expected region distribution validation error")
	}

	badDegreeDistribution := profile
	badDegreeDistribution.DegreeDistribution = []float64{0.8, 0.7, 1.0}
	if err := badDegreeDistribution.Validate(6); err == nil {
		t.Fatal("expected degree distribution validation error")
	}
}
