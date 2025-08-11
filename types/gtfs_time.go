package types

import (
	"fmt"
	"strconv"
	"strings"
)

// GTFSTime represents a time in GTFS format (HH:MM:SS or H:MM:SS)
// GTFS allows times beyond 24:00:00 for trips that span midnight
// Examples: 25:30:00 = 1:30 AM next day, 26:15:00 = 2:15 AM next day
type GTFSTime struct {
	Hours   int
	Minutes int
	Seconds int
}

// ParseGTFSTime parses a GTFS time string (HH:MM:SS or H:MM:SS)
func ParseGTFSTime(s string) (*GTFSTime, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid GTFS time format: %s (expected HH:MM:SS)", s)
	}

	// Require zero-padded minutes and seconds; hours may be unpadded per GTFS
	if len(parts[1]) != 2 || len(parts[2]) != 2 {
		return nil, fmt.Errorf("invalid zero padding in minutes/seconds: %s", s)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid hours in GTFS time: %s", parts[0])
	}
	if hours < 0 {
		return nil, fmt.Errorf("hours out of range: %d", hours)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minutes in GTFS time: %s", parts[1])
	}
	if minutes < 0 || minutes > 59 {
		return nil, fmt.Errorf("minutes out of range: %d", minutes)
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid seconds in GTFS time: %s", parts[2])
	}
	if seconds < 0 || seconds > 59 {
		return nil, fmt.Errorf("seconds out of range: %d", seconds)
	}

	return &GTFSTime{
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
	}, nil
}

// String returns the GTFS time as a string (HH:MM:SS)
func (t *GTFSTime) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hours, t.Minutes, t.Seconds)
}

// ToSeconds returns the total seconds since midnight
func (t *GTFSTime) ToSeconds() int {
	return t.Hours*3600 + t.Minutes*60 + t.Seconds
}

// Before returns true if this time is before the other time
func (t *GTFSTime) Before(other *GTFSTime) bool {
	return t.ToSeconds() < other.ToSeconds()
}

// After returns true if this time is after the other time
func (t *GTFSTime) After(other *GTFSTime) bool {
	return t.ToSeconds() > other.ToSeconds()
}

// Equal returns true if this time equals the other time
func (t *GTFSTime) Equal(other *GTFSTime) bool {
	return t.ToSeconds() == other.ToSeconds()
}
