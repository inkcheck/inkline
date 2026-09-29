// Package ids names the objects inkline generates. The vf_ prefix keeps them
// clear of author keys.
package ids

const (
	Start = "vf_start"
	End   = "vf_end"
	Chart = "vf_chart"
	XAxis = "vf_x"
	YAxis = "vf_y"
	// GapPrefix names the invisible cells that pad a lane so steps line up by rank.
	GapPrefix = "vf_gap_"
)
