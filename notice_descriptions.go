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
		"invalid_date_format": {
			Description:    "Date field contains invalid format. GTFS requires dates in YYYYMMDD format.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#field-types",
			AffectedFiles:  []string{"calendar.txt", "calendar_dates.txt", "feed_info.txt"},
			AffectedFields: []string{"start_date", "end_date", "date", "feed_start_date", "feed_end_date"},
			ExampleFix:     "Change '2023-12-25' to '20231225' or '25/12/2023' to '20231225'",
		},
		"invalid_date": {
			Description:    "A date field contains an invalid date format. Dates must be in YYYYMMDD format.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#field-types",
			AffectedFiles:  []string{"calendar.txt", "calendar_dates.txt", "feed_info.txt"},
			AffectedFields: []string{"start_date", "end_date", "date", "feed_start_date", "feed_end_date"},
			ExampleFix:     "Change '2023-12-25' to '20231225' or '25/12/2023' to '20231225'",
		},
		"invalid_time_format": {
			Description:    "Time field contains invalid format. GTFS requires times in HH:MM:SS format (24-hour clock).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#field-types",
			AffectedFiles:  []string{"stop_times.txt", "frequencies.txt"},
			AffectedFields: []string{"arrival_time", "departure_time", "start_time", "end_time"},
			ExampleFix:     "Change '2:30 PM' to '14:30:00' or '9:15' to '09:15:00'. Use '25:30:00' for next-day service.",
		},
		"invalid_coordinate": {
			Description:    "Coordinates are outside valid ranges. Latitude must be between -90 and 90, longitude between -180 and 180.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles:  []string{"stops.txt", "shapes.txt"},
			AffectedFields: []string{"stop_lat", "stop_lon", "shape_pt_lat", "shape_pt_lon"},
			ExampleFix:     "Ensure latitude is between -90 and 90 (e.g., 40.748817) and longitude is between -180 and 180 (e.g., -73.985428)",
		},
		"invalid_route_type": {
			Description:    "Route type must be a valid GTFS route type code (0-12 for basic types, 100-1799 for extended types).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles:  []string{"routes.txt"},
			AffectedFields: []string{"route_type"},
			ExampleFix:     "Use valid codes: 0=Tram, 1=Subway, 2=Rail, 3=Bus, 4=Ferry, 5=Cable, 6=Gondola, 7=Funicular, 11=Trolleybus, 12=Monorail",
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
		"same_name_and_description": {
			Description:    "Route short name and long name are identical. These should provide different levels of detail for passenger information.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles:  []string{"routes.txt"},
			AffectedFields: []string{"route_short_name", "route_long_name"},
			ExampleFix:     "Use route_short_name for '1' or 'Blue', route_long_name for 'Downtown Express'",
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
		"stop_name_missing_but_inherited": {
			Description: "Stop name is missing but can inherit from parent station. Consider adding explicit stop name for clarity.",
			ExampleFix:  "Add explicit stop_name even if it can inherit from parent_station",
		},
		"invalid_bikes_allowed": {
			Description:    "Bikes allowed field contains an invalid value. Must be 0 (no info), 1 (bikes allowed), or 2 (bikes not allowed).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#tripstxt",
			AffectedFiles:  []string{"trips.txt"},
			AffectedFields: []string{"bikes_allowed"},
			ExampleFix:     "Use 0 for no information, 1 if bikes are allowed, 2 if bikes are not allowed",
		},
		"attribution_without_role": {
			Description:    "Attribution has no role assigned. Each attribution must have at least one role (producer, operator, or authority).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#attributionstxt",
			AffectedFiles:  []string{"attributions.txt"},
			AffectedFields: []string{"is_producer", "is_operator", "is_authority"},
			ExampleFix:     "Set at least one role to 1: is_producer=1, is_operator=0, is_authority=0",
		},

		// === RELATIONSHIP VALIDATION ERRORS ===
		"duplicate_stop_sequence": {
			Description:    "Duplicate stop sequence found in a trip. Each stop in a trip must have a unique sequence number.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stop_timestxt",
			AffectedFiles:  []string{"stop_times.txt"},
			AffectedFields: []string{"stop_sequence"},
			ExampleFix:     "Ensure stop_sequence values are unique within each trip: 1, 2, 3, 4... not 1, 2, 2, 3",
		},

		// === BUSINESS LOGIC ERRORS ===
		"invalid_frequency_time_range": {
			Description:    "Frequency time range is invalid. Start time must be before end time for each frequency period.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#frequenciestxt",
			AffectedFiles:  []string{"frequencies.txt"},
			AffectedFields: []string{"start_time", "end_time"},
			ExampleFix:     "Ensure start_time < end_time: start_time=06:00:00, end_time=22:00:00",
		},
		"invalid_headway": {
			Description:    "Frequency headway is invalid. Headway must be greater than 0 seconds.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#frequenciestxt",
			AffectedFiles:  []string{"frequencies.txt"},
			AffectedFields: []string{"headway_secs"},
			ExampleFix:     "Use positive headway values: headway_secs=900 (15 minutes) instead of 0 or negative",
		},
		"overlapping_frequency": {
			Description:    "Frequency periods overlap for the same trip. Each time period should be distinct and non-overlapping.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#frequenciestxt",
			AffectedFiles:  []string{"frequencies.txt"},
			AffectedFields: []string{"start_time", "end_time", "trip_id"},
			ExampleFix:     "Ensure non-overlapping periods: Period 1: 06:00-12:00, Period 2: 12:00-18:00",
		},
		"invalid_transfer_type": {
			Description:    "Transfer type is invalid. Must be 0 (recommended), 1 (timed), 2 (minimum time), or 3 (not possible).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#transferstxt",
			AffectedFiles:  []string{"transfers.txt"},
			AffectedFields: []string{"transfer_type"},
			ExampleFix:     "Use valid values: 0=recommended, 1=timed, 2=minimum time required, 3=not possible",
		},
		"expired_feed": {
			Description:    "The feed has expired. Feeds should be updated regularly to provide current service information.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#feed_infotxt",
			AffectedFiles:  []string{"feed_info.txt"},
			AffectedFields: []string{"feed_end_date"},
			ExampleFix:     "Update feed_end_date to a current date: feed_end_date=20241231",
		},

		// === ACCESSIBILITY ERRORS ===
		"invalid_pathway_mode": {
			Description:    "Pathway mode is invalid or not specified. Pathway modes must be valid GTFS pathway type codes (1-7).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#pathwaystxt",
			AffectedFiles:  []string{"pathways.txt"},
			AffectedFields: []string{"pathway_mode"},
			ExampleFix:     "Use valid codes: 1=walkway, 2=stairs, 3=moving_sidewalk, 4=escalator, 5=elevator, 6=fare_gate, 7=exit_gate",
		},

		// === FARE SYSTEM ERRORS ===
		"invalid_fare_price": {
			Description:    "Fare price is negative or has excessive precision. Prices should be non-negative with reasonable decimal places.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_attributestxt",
			AffectedFiles:  []string{"fare_attributes.txt"},
			AffectedFields: []string{"price"},
			ExampleFix:     "Use valid prices: price=2.50 (not -2.50 or 2.123456)",
		},
		"invalid_payment_method": {
			Description:    "Payment method is invalid. Must be 0 (paid on board) or 1 (paid before boarding).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_attributestxt",
			AffectedFiles:  []string{"fare_attributes.txt"},
			AffectedFields: []string{"payment_method"},
			ExampleFix:     "Use 0 for pay-on-board or 1 for prepaid tickets/cards",
		},
		"empty_fare_rule": {
			Description:    "A fare rule has no qualifying conditions specified. At least one condition must be defined.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_rulestxt",
			AffectedFiles:  []string{"fare_rules.txt"},
			AffectedFields: []string{"route_id", "origin_id", "destination_id", "contains_id"},
			ExampleFix:     "Specify at least one condition: route_id=R1 or origin_id=zone_A",
		},

		// === GEOGRAPHIC DATA ERRORS ===
		"suspicious_coordinate": {
			Description:    "Coordinates appear to be placeholder or error values (e.g., 0,0). This may indicate data import issues.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles:  []string{"stops.txt", "shapes.txt"},
			AffectedFields: []string{"stop_lat", "stop_lon", "shape_pt_lat", "shape_pt_lon"},
			ExampleFix:     "Replace with actual coordinates: stop_lat=40.748817, stop_lon=-73.985428",
		},

		// === NETWORK TOPOLOGY ERRORS ===

		// === FEED INFO ERRORS ===

		// === TRIP AND SERVICE ERRORS ===
		"service_never_active": {
			Description:    "A service is defined but never active on any day. This creates unused service definitions.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#calendartxt",
			AffectedFiles:  []string{"calendar.txt"},
			AffectedFields: []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"},
			ExampleFix:     "Set at least one day to 1: monday=1, or remove unused service",
		},
		"unused_service": {
			Description:    "A service is defined but not used by any trips. This creates orphaned service definitions.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#tripstxt",
			AffectedFiles:  []string{"calendar.txt", "trips.txt"},
			AffectedFields: []string{"service_id"},
			ExampleFix:     "Remove unused services or add trips that reference them",
		},
		"invalid_currency_code": {
			Description:    "Currency code is invalid or not recognized. Must be a valid ISO 4217 3-letter currency code.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_attributestxt",
			AffectedFiles:  []string{"fare_attributes.txt"},
			AffectedFields: []string{"currency_type"},
			ExampleFix:     "Use valid codes: USD, EUR, CAD, GBP, JPY, etc.",
		},
		"validator_error": {
			Description: "A validator encountered an error during processing. This may indicate data corruption or validator issues.",
			ExampleFix:  "Check data file integrity and report issue if problem persists",
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
