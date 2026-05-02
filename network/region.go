package network

// RegionID identifies a geographic region in the network model.
type RegionID int

const (
	RegionNorthAmerica RegionID = iota
	RegionEurope
	RegionSouthAmerica
	RegionAsiaPacific
	RegionJapan
	RegionAustralia
)

// Region captures static metadata for one geographic region.
type Region struct {
	ID   RegionID
	Name string
}

// DefaultRegions mirrors SimBlock's built-in region list.
var DefaultRegions = []Region{
	{ID: RegionNorthAmerica, Name: "NORTH_AMERICA"},
	{ID: RegionEurope, Name: "EUROPE"},
	{ID: RegionSouthAmerica, Name: "SOUTH_AMERICA"},
	{ID: RegionAsiaPacific, Name: "ASIA_PACIFIC"},
	{ID: RegionJapan, Name: "JAPAN"},
	{ID: RegionAustralia, Name: "AUSTRALIA"},
}

// DefaultRegionNames returns region names in stable ID order.
func DefaultRegionNames() []string {
	out := make([]string, 0, len(DefaultRegions))
	for _, region := range DefaultRegions {
		out = append(out, region.Name)
	}
	return out
}
