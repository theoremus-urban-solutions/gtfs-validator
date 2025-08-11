package types

import (
	"fmt"
	"strconv"
	"strings"
)

// GTFSColor represents a color in GTFS format (6-digit hexadecimal)
type GTFSColor struct {
	R, G, B uint8
}

// ParseGTFSColor parses a GTFS color string (6-digit hex without #)
func ParseGTFSColor(s string) (*GTFSColor, error) {
	// Remove # if present
	s = strings.TrimPrefix(s, "#")

	if len(s) != 6 {
		return nil, fmt.Errorf("invalid GTFS color format: %s (expected 6 hex digits)", s)
	}

	rgb, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", s)
	}

	return &GTFSColor{
		R: uint8(rgb >> 16),
		G: uint8(rgb >> 8),
		B: uint8(rgb),
	}, nil
}

// String returns the GTFS color as a 6-digit hex string
func (c *GTFSColor) String() string {
	return fmt.Sprintf("%02X%02X%02X", c.R, c.G, c.B)
}

// ToHTMLColor returns the color as an HTML color string (with #)
func (c *GTFSColor) ToHTMLColor() string {
	return "#" + c.String()
}

// Luminance calculates the relative luminance of the color
// Used for determining color contrast
func (c *GTFSColor) Luminance() float64 {
	// Convert to linear RGB
	r := c.linearize(float64(c.R) / 255.0)
	g := c.linearize(float64(c.G) / 255.0)
	b := c.linearize(float64(c.B) / 255.0)

	// Calculate relative luminance
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// linearize converts sRGB to linear RGB
func (c *GTFSColor) linearize(channel float64) float64 {
	if channel <= 0.03928 {
		return channel / 12.92
	}
	return ((channel + 0.055) / 1.055) * ((channel + 0.055) / 1.055)
}

// ContrastRatio calculates the contrast ratio between two colors
// Returns a value between 1 and 21 (higher is better contrast)
func (c *GTFSColor) ContrastRatio(other *GTFSColor) float64 {
	l1 := c.Luminance()
	l2 := other.Luminance()

	if l1 > l2 {
		return (l1 + 0.05) / (l2 + 0.05)
	}
	return (l2 + 0.05) / (l1 + 0.05)
}
