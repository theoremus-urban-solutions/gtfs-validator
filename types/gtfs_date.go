package types

import (
	"fmt"
	"strconv"
	"time"
)

// GTFSDate represents a date in GTFS format (YYYYMMDD)
type GTFSDate struct {
	Year  int
	Month int
	Day   int
}

// ParseGTFSDate parses a GTFS date string (YYYYMMDD)
func ParseGTFSDate(s string) (*GTFSDate, error) {
	if len(s) != 8 {
		return nil, fmt.Errorf("invalid GTFS date format: %s (expected YYYYMMDD)", s)
	}

	year, err := strconv.Atoi(s[0:4])
	if err != nil {
		return nil, fmt.Errorf("invalid year in GTFS date: %s", s[0:4])
	}

	month, err := strconv.Atoi(s[4:6])
	if err != nil {
		return nil, fmt.Errorf("invalid month in GTFS date: %s", s[4:6])
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month out of range: %d", month)
	}

	day, err := strconv.Atoi(s[6:8])
	if err != nil {
		return nil, fmt.Errorf("invalid day in GTFS date: %s", s[6:8])
	}
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day out of range: %d", day)
	}

	// Validate the date
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Day() != day || t.Month() != time.Month(month) || t.Year() != year {
		return nil, fmt.Errorf("invalid date: %s", s)
	}

	return &GTFSDate{
		Year:  year,
		Month: month,
		Day:   day,
	}, nil
}

// String returns the GTFS date as a string (YYYYMMDD)
func (d *GTFSDate) String() string {
	return fmt.Sprintf("%04d%02d%02d", d.Year, d.Month, d.Day)
}

// ToTime converts the GTFS date to a time.Time
func (d *GTFSDate) ToTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// Before returns true if this date is before the other date
func (d *GTFSDate) Before(other *GTFSDate) bool {
	if d.Year != other.Year {
		return d.Year < other.Year
	}
	if d.Month != other.Month {
		return d.Month < other.Month
	}
	return d.Day < other.Day
}

// After returns true if this date is after the other date
func (d *GTFSDate) After(other *GTFSDate) bool {
	return other.Before(d)
}

// Equal returns true if this date equals the other date
func (d *GTFSDate) Equal(other *GTFSDate) bool {
	return d.Year == other.Year && d.Month == other.Month && d.Day == other.Day
}