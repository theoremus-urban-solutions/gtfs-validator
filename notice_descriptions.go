package gtfsvalidator

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
)

// NoticeDescription contains enhanced information about a validation notice
type NoticeDescription struct {
	Description    string   `json:"description"`
	GTFSReference  string   `json:"gtfsReference,omitempty"`
	AffectedFiles  []string `json:"affectedFiles,omitempty"`
	AffectedFields []string `json:"affectedFields,omitempty"`
	ExampleFix     string   `json:"exampleFix,omitempty"`
}

// SeverityInfo provides detailed information about validation severity levels
type SeverityInfo struct {
	Level       string `json:"level"`
	Description string `json:"description"`
	Urgency     string `json:"urgency"`
}

// severityDescriptions maps severity levels to their detailed information
var severityDescriptions = map[string]SeverityInfo{
	"ERROR": {
		Level:       "ERROR",
		Description: "Critical GTFS compliance violation",
		Urgency:     "Must fix before publishing feed",
	},
	"WARNING": {
		Level:       "WARNING",
		Description: "Data quality issue that may affect user experience",
		Urgency:     "Should fix to improve feed quality",
	},
	"INFO": {
		Level:       "INFO",
		Description: "Best practice recommendation or informational notice",
		Urgency:     "Consider fixing for optimal feed quality",
	},
}

// GetSeverityInfo returns detailed information about a severity level
func GetSeverityInfo(severity string) SeverityInfo {
	if info, exists := severityDescriptions[strings.ToUpper(severity)]; exists {
		return info
	}
	return SeverityInfo{
		Level:       severity,
		Description: "Unknown severity level",
		Urgency:     "Review validation result",
	}
}

// affectedFiles returns the GTFS files a notice code concerns. Only a minority
// of codes have a hand-written description, so fall back to the per-code file
// mapping, which covers every code that is tied to a specific file.
func affectedFiles(code string, enhanced NoticeDescription) []string {
	if len(enhanced.AffectedFiles) > 0 {
		return enhanced.AffectedFiles
	}
	return notice.AffectedFilesForCode(code)
}

// getNoticeDescription returns a comprehensive, user-friendly description for a notice code
// These descriptions help feed producers understand and fix validation issues
func getNoticeDescription(code string) string {
	enhanced := GetEnhancedNoticeDescription(code)
	return enhanced.Description
}

// GetEnhancedNoticeDescription returns detailed notice information including GTFS references
func GetEnhancedNoticeDescription(code string) NoticeDescription {
	descriptions := map[string]NoticeDescription{
		// === CORE VALIDATION ERRORS ===
		"missing_required_file": {
			Description:   "A required GTFS file is missing from the feed. This file is essential for GTFS compliance and must be present.",
			GTFSReference: "https://gtfs.org/schedule/reference/#dataset-files",
			AffectedFiles: []string{"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt"},
			ExampleFix:    "Create the missing file with required headers and data. For example, agency.txt must contain: agency_id,agency_name,agency_url,agency_timezone",
		},
		"missing_required_field": {
			Description:   "A required field is missing from a GTFS file. This field is mandatory according to the GTFS specification.",
			GTFSReference: "https://gtfs.org/schedule/reference/#field-definitions",
			ExampleFix:    "Add the missing field to the file header and provide values for all rows. For example, add 'stop_name' column to stops.txt",
		},
		"empty_file": {
			Description:   "A GTFS file is completely empty (no data rows). Empty files may indicate data export issues or missing content.",
			GTFSReference: "https://gtfs.org/schedule/reference/#dataset-files",
			ExampleFix:    "Remove the empty file if not needed, or add proper header and data rows",
		},
		"invalid_date": {
			Description:    "A date field contains an invalid date format. Dates must be in YYYYMMDD format.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#field-types",
			AffectedFiles:  []string{"calendar.txt", "calendar_dates.txt", "feed_info.txt"},
			AffectedFields: []string{"start_date", "end_date", "date", "feed_start_date", "feed_end_date"},
			ExampleFix:     "Change '2023-12-25' to '20231225' or '25/12/2023' to '20231225'",
		},
		"duplicate_key": {
			Description:   "A record has a duplicate primary key. Each record must have a unique identifier to maintain data integrity.",
			GTFSReference: "https://gtfs.org/schedule/reference/#field-definitions",
			ExampleFix:    "Ensure each record has a unique primary key value. For stops.txt, each stop_id must be unique.",
		},

		// === ENTITY VALIDATION ERRORS ===
		"route_both_short_and_long_name_missing": {
			Description:    "Both route_short_name and route_long_name are empty. At least one route name must be provided for passenger identification.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles:  []string{"routes.txt"},
			AffectedFields: []string{"route_short_name", "route_long_name"},
			ExampleFix:     "Add either route_short_name (e.g., '1', 'Blue Line') or route_long_name (e.g., 'Downtown Express')",
		},
		"duplicate_route_name": {
			Description:   "Multiple routes have the same name. This may cause confusion for passengers and should be differentiated.",
			GTFSReference: "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles: []string{"routes.txt"},
			ExampleFix:    "Ensure each route has a unique name combination of short_name and long_name",
		},
		"missing_stop_name": {
			Description:    "A required stop name is missing. Stop names are essential for passenger identification and navigation.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles:  []string{"stops.txt"},
			AffectedFields: []string{"stop_name"},
			ExampleFix:     "Add descriptive stop names like 'Main St & 1st Ave' or 'Downtown Transit Center'",
		},
		"foreign_key_violation": {
			Description:   "A foreign key reference is invalid. The referenced record does not exist in the target file.",
			GTFSReference: "https://gtfs.org/schedule/reference/#field-definitions",
			ExampleFix:    "Ensure referenced IDs exist. For example, if trips.txt references route_id 'R1', ensure 'R1' exists in routes.txt",
		},
		"fast_travel_between_consecutive_stops": {
			Description:    "Travel speed between stops is unrealistically fast for the transport mode. This may indicate data errors or missing stops.",
			GTFSReference:  "https://gtfs.org/schedule/best-practices/#stop_timestxt",
			AffectedFiles:  []string{"stop_times.txt"},
			AffectedFields: []string{"arrival_time", "departure_time"},
			ExampleFix:     "Check for missing intermediate stops or correct travel times. Bus speeds should typically be under 100 km/h.",
		},
		"attribution_without_role": {
			Description:    "Attribution has no role assigned. Each attribution must have at least one role (producer, operator, or authority).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#attributionstxt",
			AffectedFiles:  []string{"attributions.txt"},
			AffectedFields: []string{"is_producer", "is_operator", "is_authority"},
			ExampleFix:     "Set at least one role to 1: is_producer=1, is_operator=0, is_authority=0",
		},

		// === RELATIONSHIP VALIDATION ERRORS ===

		// === BUSINESS LOGIC ERRORS ===
		"overlapping_frequency": {
			Description:    "Frequency periods overlap for the same trip. Each time period should be distinct and non-overlapping.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#frequenciestxt",
			AffectedFiles:  []string{"frequencies.txt"},
			AffectedFields: []string{"start_time", "end_time", "trip_id"},
			ExampleFix:     "Ensure non-overlapping periods: Period 1: 06:00-12:00, Period 2: 12:00-18:00",
		},

		// === ACCESSIBILITY ERRORS ===

		// === FARE SYSTEM ERRORS ===
		"empty_fare_rule": {
			Description:    "A fare rule has no qualifying conditions specified. At least one condition must be defined.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_rulestxt",
			AffectedFiles:  []string{"fare_rules.txt"},
			AffectedFields: []string{"route_id", "origin_id", "destination_id", "contains_id"},
			ExampleFix:     "Specify at least one condition: route_id=R1 or origin_id=zone_A",
		},

		// === GEOGRAPHIC DATA ERRORS ===

		// === NETWORK TOPOLOGY ERRORS ===

		// === FEED INFO ERRORS ===

		// === TRIP AND SERVICE ERRORS ===
		"unused_service": {
			Description:    "A service is defined but not used by any trips. This creates orphaned service definitions.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#tripstxt",
			AffectedFiles:  []string{"calendar.txt", "trips.txt"},
			AffectedFields: []string{"service_id"},
			ExampleFix:     "Remove unused services or add trips that reference them",
		},
		"runtime_exception_in_validator_error": {
			Description: "A validator panicked and its checks did not run, so the report is incomplete for the files it covers. This is a bug in the validator, not a defect in the feed.",
			ExampleFix:  "Report the issue with the feed that triggered it; the remaining validators' results are still valid",
		},
	}

	desc := descriptions[code]

	// MobilityData publishes the canonical wording for every code the Canonical
	// GTFS Schedule Validator defines, so prefer it over anything hand-written
	// here. Regenerate with scripts/gen_notice_descriptions.py.
	if upstream, ok := mobilityDataDescriptions[code]; ok {
		desc.Description = upstream
		return desc
	}
	if desc.Description != "" {
		return desc
	}

	// Codes with no description anywhere fall back to the code name alone. It
	// says nothing the code does not, but a sentence that applies equally to
	// every code says less.
	desc.Description = titleFromCode(code)
	return desc
}

// titleFromCode turns a notice code into a readable title, so that "loop_route"
// becomes "Loop Route".
func titleFromCode(code string) string {
	words := strings.Split(code, "_")
	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word)
	}
	return strings.Join(words, " ")
}
