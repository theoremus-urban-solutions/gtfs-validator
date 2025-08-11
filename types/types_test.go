package types

import "testing"

func TestParseGTFSTime_Basic(t *testing.T) {
	cases := []struct {
		in      string
		ok      bool
		seconds int
	}{
		{"00:00:00", true, 0},
		{"08:05:09", true, 8*3600 + 5*60 + 9},
		{"24:00:00", true, 24 * 3600},       // allowed by GTFS
		{"25:00:00", true, 25 * 3600},       // late night service (1 AM next day)
		{"25:30:00", true, 25*3600 + 30*60}, // 1:30 AM next day
		{"26:45:00", true, 26*3600 + 45*60}, // 2:45 AM next day
		{"8:00:00", true, 8 * 3600},         // hours may be unpadded per GTFS
		{"08:0:00", false, 0},               // invalid padding
		{"08:00:0", false, 0},               // invalid padding
	}
	for _, c := range cases {
		tt, err := ParseGTFSTime(c.in)
		if c.ok && err != nil {
			t.Errorf("expected ok for %s, got err %v", c.in, err)
			continue
		}
		if !c.ok && err == nil {
			t.Errorf("expected error for %s", c.in)
			continue
		}
		if c.ok {
			if got := tt.ToSeconds(); got != c.seconds {
				t.Errorf("%s seconds expected %d, got %d", c.in, c.seconds, got)
			}
		}
	}
}

func TestParseGTFSDate_Basic(t *testing.T) {
	cases := []struct {
		in      string
		ok      bool
		y, m, d int
	}{
		{"20250101", true, 2025, 1, 1},
		{"20240229", true, 2024, 2, 29}, // leap day
		{"20250230", false, 0, 0, 0},    // invalid day
		{"2025-01-01", false, 0, 0, 0},  // wrong format
		{"2025010", false, 0, 0, 0},     // wrong length
	}
	for _, c := range cases {
		d, err := ParseGTFSDate(c.in)
		if c.ok && err != nil {
			t.Errorf("expected ok for %s, got err %v", c.in, err)
			continue
		}
		if !c.ok && err == nil {
			t.Errorf("expected error for %s", c.in)
			continue
		}
		if c.ok {
			if d.Year != c.y || d.Month != c.m || d.Day != c.d {
				t.Errorf("%s parsed mismatch got %04d-%02d-%02d", c.in, d.Year, d.Month, d.Day)
			}
		}
	}
}
