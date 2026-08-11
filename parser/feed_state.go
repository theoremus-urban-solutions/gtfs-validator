package parser

import (
	"io"
	"strings"
	"sync"
)

// FileState says whether a file is fit to be depended on by other checks.
//
// A GTFS feed is a set of tables joined by id. When one of those tables cannot
// be loaded — it is absent, or empty, or its id column is gone, or its ids are
// blank — every reference into it dangles. Reporting each of those references
// turns one defect into thousands: an emptied stops.txt produced 4,043
// foreign key violations where the canonical validator reports one empty file.
// Checks that read a table therefore ask its state first and stand down when
// the answer is anything but FileStateParsed, so the defect is reported once,
// at its cause.
type FileState int

const (
	// FileStateParsed means the file is present, readable, and every row
	// carries the key other files reference it by.
	FileStateParsed FileState = iota

	// FileStateMissing means the feed does not contain the file.
	FileStateMissing

	// FileStateEmpty means the file has a header but no data rows, or no
	// content at all.
	FileStateEmpty

	// FileStateUnparseable means the file could not be read as CSV.
	FileStateUnparseable

	// FileStateMissingKeyColumn means the column other files join on is not
	// among the headers, so no row can be referenced.
	FileStateMissingKeyColumn

	// FileStateInvalidKeyValues means at least one row leaves the key blank.
	// The row cannot be referenced and is excluded from lookups; the blank
	// itself is reported as a missing required field.
	FileStateInvalidKeyValues
)

// Usable reports whether a reference into this file can be judged. A single
// blank key is enough to make it false: the rows pointing at that key dangle
// for a reason already reported against the row that lost its id, and
// reporting them too restates one defect once per referencing row.
func (s FileState) Usable() bool { return s == FileStateParsed }

// LoadFailed reports whether the file failed to load as a table at all, as
// opposed to loading with a bad row in it.
//
// This is the stricter test, and the one that stands a whole validator down. A
// blank key on one row does not qualify: every other row is intact, and a check
// that examines rows independently — whether a zone is used, whether a name is
// mixed case — still has something true to say about them. Only a table that
// produced no usable rows at all silences its dependants.
//
// A missing file does not qualify either. Its absence is reported once as a
// missing file if it was required, and a validator that reads an absent file
// finds nothing and reports nothing of its own accord; standing it down would
// add a skip record for a file the feed never claimed to have.
func (s FileState) LoadFailed() bool {
	switch s {
	case FileStateEmpty, FileStateUnparseable, FileStateMissingKeyColumn:
		return true
	default:
		return false
	}
}

// Reason is a short phrase naming the defect, for the skip record.
func (s FileState) Reason() string {
	switch s {
	case FileStateParsed:
		return "parsed"
	case FileStateMissing:
		return "missing"
	case FileStateEmpty:
		return "empty"
	case FileStateUnparseable:
		return "unparseable"
	case FileStateMissingKeyColumn:
		return "missing its key column"
	case FileStateInvalidKeyValues:
		return "has rows with a blank key"
	default:
		return "unknown"
	}
}

// keyColumns are the columns a file is joined on from elsewhere. Only files
// something references are listed: a table nothing looks up by id cannot strand
// a reference, whatever state it is in, and reading it here would cost a pass
// over the largest file in the feed for nothing.
//
// agency.txt is deliberately absent. Its agency_id is optional in a feed with a
// single agency, so a blank one is not a defect and must not stand anything down.
var keyColumns = map[string]string{
	"stops.txt":           "stop_id",
	"routes.txt":          "route_id",
	"trips.txt":           "trip_id",
	"calendar.txt":        "service_id",
	"calendar_dates.txt":  "service_id",
	"shapes.txt":          "shape_id",
	"levels.txt":          "level_id",
	"pathways.txt":        "pathway_id",
	"fare_attributes.txt": "fare_id",
}

// AbsenceStrandsReferences reports whether references into a file should still
// be judged when the file is not present at all.
//
// This encodes a canonical asymmetry rather than a principle. A file the spec
// requires outright — routes, trips, stop_times, agency — being absent stops
// the feed loading as a dataset, and canonical answers with the missing-file
// error alone rather than a violation per reference. stops.txt is only
// conditionally required since locations.geojson arrived, so canonical loads the
// feed without it and does report every stop reference as dangling: 4,043 of
// them on a small feed. Matching canonical means copying that distinction.
func AbsenceStrandsReferences(filename string) bool {
	return filename == "stops.txt"
}

// FileState returns the state of one file, computing it on first request and
// remembering the answer. Several validators ask about the same file, and the
// answer cannot change during a validation run.
func (l *FeedLoader) FileState(filename string) FileState {
	l.stateOnce.Do(func() {
		l.states = make(map[string]FileState, len(keyColumns))
		l.stateMu = &sync.Mutex{}
	})

	l.stateMu.Lock()
	defer l.stateMu.Unlock()
	if state, known := l.states[filename]; known {
		return state
	}
	state := l.computeFileState(filename)
	l.states[filename] = state
	return state
}

// computeFileState reads the file far enough to classify it.
func (l *FeedLoader) computeFileState(filename string) FileState {
	if !l.HasFile(filename) {
		return FileStateMissing
	}

	reader, err := l.GetFile(filename)
	if err != nil {
		return FileStateMissing
	}
	defer func() { _ = reader.Close() }()

	csvFile, err := NewCSVFile(reader, filename)
	if err != nil {
		if strings.Contains(err.Error(), "empty file") {
			return FileStateEmpty
		}
		return FileStateUnparseable
	}

	key, joined := keyColumns[filename]
	if joined {
		found := false
		for _, header := range csvFile.Headers {
			if strings.TrimSpace(header) == key {
				found = true
				break
			}
		}
		if !found {
			return FileStateMissingKeyColumn
		}
	}

	blankKeys := false
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			// A row the CSV reader rejects is reported by the structural
			// checks; it does not by itself make the table unusable.
			continue
		}
		if joined && strings.TrimSpace(row.Values[key]) == "" {
			blankKeys = true
		}
	}

	// A table with a valid header and no data rows is NOT empty — it is a
	// legitimately empty table that loaded correctly, and it suppresses nothing.
	// Canonical draws exactly this line: a zero-byte file is EMPTY_FILE and
	// stands its dependants down, while a header with no rows loads fine and
	// every reference into it is reported as the violation it is. Collapsing the
	// two hid 4,043 canonical foreign key ERRORs on a feed with an emptied
	// stops.txt.
	if blankKeys {
		return FileStateInvalidKeyValues
	}
	return FileStateParsed
}
