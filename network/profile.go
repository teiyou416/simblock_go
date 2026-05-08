package network

import "fmt"

// Profile describes a reusable network profile for region metadata,
// bandwidth, and node/topology distributions.
type Profile struct {
	Name               string
	Regions            []Region
	UploadBandwidth    []uint64
	DownloadBandwidth  []uint64
	RegionDistribution []float64
	DegreeDistribution []float64
}

// Built-in Java SimBlock Bitcoin 2019/2015 network profile.
//
// Regions, latency, bandwidth, and region distribution mirror Java's
// NetworkConfiguration 2019 values. Degree distribution mirrors the Java
// Bitcoin 2015 outbound-link cumulative distribution.
var Bitcoin2019Profile = Profile{
	Name:    "bitcoin_2019",
	Regions: DefaultRegions,
	UploadBandwidth: []uint64{
		19_200_000,
		20_700_000,
		5_800_000,
		15_700_000,
		10_200_000,
		11_300_000,
	},
	DownloadBandwidth: []uint64{
		52_000_000,
		40_000_000,
		18_000_000,
		22_800_000,
		22_800_000,
		29_900_000,
	},
	RegionDistribution: []float64{
		0.3316,
		0.4998,
		0.0090,
		0.1177,
		0.0224,
		0.0195,
	},
	DegreeDistribution: []float64{
		0.025,
		0.050,
		0.075,
		0.10,
		0.20,
		0.30,
		0.40,
		0.50,
		0.60,
		0.70,
		0.80,
		0.85,
		0.90,
		0.95,
		0.97,
		0.97,
		0.98,
		0.99,
		0.995,
		1.0,
	},
}

// ProfileByName returns a built-in profile by name.
func ProfileByName(name string) (Profile, bool) {
	switch name {
	case "", Bitcoin2019Profile.Name:
		return Bitcoin2019Profile, true
	default:
		return Profile{}, false
	}
}

// ProfileOverrides contains optional config-level overrides for a base profile.
type ProfileOverrides struct {
	UploadBandwidth    []uint64
	DownloadBandwidth  []uint64
	RegionDistribution []float64
	DegreeDistribution []float64
}

// WithOverrides returns a profile with explicit config values applied.
func (p Profile) WithOverrides(overrides ProfileOverrides) Profile {
	if len(overrides.UploadBandwidth) > 0 {
		p.UploadBandwidth = append([]uint64(nil), overrides.UploadBandwidth...)
	}
	if len(overrides.DownloadBandwidth) > 0 {
		p.DownloadBandwidth = append([]uint64(nil), overrides.DownloadBandwidth...)
	}
	if len(overrides.RegionDistribution) > 0 {
		p.RegionDistribution = append([]float64(nil), overrides.RegionDistribution...)
	}
	if len(overrides.DegreeDistribution) > 0 {
		p.DegreeDistribution = append([]float64(nil), overrides.DegreeDistribution...)
	}
	return p
}

// Validate checks that a profile can be used with a latency matrix of regionCount regions.
func (p Profile) Validate(regionCount int) error {
	if regionCount <= 0 {
		return fmt.Errorf("region count must be positive")
	}
	if len(p.UploadBandwidth) == 0 {
		return fmt.Errorf("upload bandwidth is empty")
	}
	if len(p.DownloadBandwidth) == 0 {
		return fmt.Errorf("download bandwidth is empty")
	}
	if len(p.UploadBandwidth) != regionCount {
		return fmt.Errorf("upload bandwidth length %d does not match region count %d", len(p.UploadBandwidth), regionCount)
	}
	if len(p.DownloadBandwidth) != regionCount {
		return fmt.Errorf("download bandwidth length %d does not match region count %d", len(p.DownloadBandwidth), regionCount)
	}
	if err := validateDistribution("region distribution", p.RegionDistribution, regionCount, false); err != nil {
		return err
	}
	if err := validateDistribution("degree distribution", p.DegreeDistribution, 1, true); err != nil {
		return err
	}
	return nil
}

// Bandwidths returns profile bandwidth arrays sized for regionCount.
func (p Profile) Bandwidths(regionCount int) ([]uint64, []uint64) {
	upload := repeatToLength(p.UploadBandwidth, regionCount)
	download := repeatToLength(p.DownloadBandwidth, regionCount)
	return upload, download
}

func validateDistribution(name string, values []float64, minLen int, cumulative bool) error {
	if len(values) < minLen {
		return fmt.Errorf("%s length %d is less than %d", name, len(values), minLen)
	}
	prev := 0.0
	sum := 0.0
	for i, value := range values {
		if value < 0 || value > 1 {
			return fmt.Errorf("%s value at %d must be in [0,1], got %f", name, i, value)
		}
		if cumulative && value < prev {
			return fmt.Errorf("%s must be non-decreasing", name)
		}
		prev = value
		sum += value
	}
	if cumulative {
		if values[len(values)-1] != 1.0 {
			return fmt.Errorf("%s final value must be 1.0", name)
		}
		return nil
	}
	if sum < 0.999 || sum > 1.001 {
		return fmt.Errorf("%s must sum to 1.0, got %f", name, sum)
	}
	return nil
}

func repeatToLength(values []uint64, length int) []uint64 {
	out := make([]uint64, length)
	if len(values) == 0 {
		return out
	}
	for i := 0; i < length; i++ {
		out[i] = values[i%len(values)]
	}
	return out
}
