package gtfsvalidator

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NoticeDescription contains enhanced information about a validation notice
type NoticeDescription struct {
	Description    string   `json:"description"`
	GTFSReference  string   `json:"gtfsReference,omitempty"`
	AffectedFiles  []string `json:"affectedFiles,omitempty"`
	AffectedFields []string `json:"affectedFields,omitempty"`
	ExampleFix     string   `json:"exampleFix,omitempty"`
	Impact         string   `json:"impact,omitempty"`
}

// SeverityInfo provides detailed information about validation severity levels
type SeverityInfo struct {
	Level       string `json:"level"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Urgency     string `json:"urgency"`
}

// severityDescriptions maps severity levels to their detailed information
var severityDescriptions = map[string]SeverityInfo{
	"ERROR": {
		Level:       "ERROR",
		Description: "Critical GTFS compliance violation",
		Impact:      "Feed may be rejected by transit apps and journey planners",
		Urgency:     "Must fix before publishing feed",
	},
	"WARNING": {
		Level:       "WARNING",
		Description: "Data quality issue that may affect user experience",
		Impact:      "May cause confusion for passengers or reduced functionality",
		Urgency:     "Should fix to improve feed quality",
	},
	"INFO": {
		Level:       "INFO",
		Description: "Best practice recommendation or informational notice",
		Impact:      "Helps optimize feed for better passenger experience",
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
		Impact:      "Impact unknown",
		Urgency:     "Review validation result",
	}
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
			Impact:        "Feed will not be accepted by GTFS consumers and transit apps",
			ExampleFix:    "Create the missing file with required headers and data. For example, agency.txt must contain: agency_id,agency_name,agency_url,agency_timezone",
		},
		"missing_required_field": {
			Description:   "A required field is missing from a GTFS file. This field is mandatory according to the GTFS specification.",
			GTFSReference: "https://gtfs.org/schedule/reference/#field-definitions",
			Impact:        "Data integrity issues, potential feed rejection by transit applications",
			ExampleFix:    "Add the missing field to the file header and provide values for all rows. For example, add 'stop_name' column to stops.txt",
		},
		"empty_file": {
			Description:   "A GTFS file is completely empty (no data rows). Empty files may indicate data export issues or missing content.",
			GTFSReference: "https://gtfs.org/schedule/reference/#dataset-files",
			Impact:        "File parsing errors, incomplete feed information",
			ExampleFix:    "Remove the empty file if not needed, or add proper header and data rows",
		},
		"invalid_date_format": {
			Description:    "Date field contains invalid format. GTFS requires dates in YYYYMMDD format.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#field-types",
			AffectedFiles:  []string{"calendar.txt", "calendar_dates.txt", "feed_info.txt"},
			AffectedFields: []string{"start_date", "end_date", "date", "feed_start_date", "feed_end_date"},
			Impact:         "Date parsing errors, service scheduling issues",
			ExampleFix:     "Change '2023-12-25' to '20231225' or '25/12/2023' to '20231225'",
		},
		"invalid_date": {
			Description:    "A date field contains an invalid date format. Dates must be in YYYYMMDD format.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#field-types",
			AffectedFiles:  []string{"calendar.txt", "calendar_dates.txt", "feed_info.txt"},
			AffectedFields: []string{"start_date", "end_date", "date", "feed_start_date", "feed_end_date"},
			Impact:         "Date parsing errors, service scheduling issues",
			ExampleFix:     "Change '2023-12-25' to '20231225' or '25/12/2023' to '20231225'",
		},
		"invalid_time_format": {
			Description:    "Time field contains invalid format. GTFS requires times in HH:MM:SS format (24-hour clock).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#field-types",
			AffectedFiles:  []string{"stop_times.txt", "frequencies.txt"},
			AffectedFields: []string{"arrival_time", "departure_time", "start_time", "end_time"},
			Impact:         "Time parsing errors, trip scheduling issues",
			ExampleFix:     "Change '2:30 PM' to '14:30:00' or '9:15' to '09:15:00'. Use '25:30:00' for next-day service.",
		},
		"invalid_coordinate": {
			Description:    "Coordinates are outside valid ranges. Latitude must be between -90 and 90, longitude between -180 and 180.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles:  []string{"stops.txt", "shapes.txt"},
			AffectedFields: []string{"stop_lat", "stop_lon", "shape_pt_lat", "shape_pt_lon"},
			Impact:         "Mapping errors, location display issues",
			ExampleFix:     "Ensure latitude is between -90 and 90 (e.g., 40.748817) and longitude is between -180 and 180 (e.g., -73.985428)",
		},
		"invalid_route_type": {
			Description:    "Route type must be a valid GTFS route type code (0-12 for basic types, 100-1799 for extended types).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles:  []string{"routes.txt"},
			AffectedFields: []string{"route_type"},
			Impact:         "Route classification errors, consumer confusion",
			ExampleFix:     "Use valid codes: 0=Tram, 1=Subway, 2=Rail, 3=Bus, 4=Ferry, 5=Cable, 6=Gondola, 7=Funicular, 11=Trolleybus, 12=Monorail",
		},
		"duplicate_key": {
			Description:   "A record has a duplicate primary key. Each record must have a unique identifier to maintain data integrity.",
			GTFSReference: "https://gtfs.org/schedule/reference/#field-definitions",
			Impact:        "Data conflicts, potential feed rejection",
			ExampleFix:    "Ensure each record has a unique primary key value. For stops.txt, each stop_id must be unique.",
		},

		// === ENTITY VALIDATION ERRORS ===
		"missing_route_name": {
			Description:    "Both route_short_name and route_long_name are empty. At least one route name must be provided for passenger identification.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles:  []string{"routes.txt"},
			AffectedFields: []string{"route_short_name", "route_long_name"},
			Impact:         "Passengers cannot identify routes, poor user experience",
			ExampleFix:     "Add either route_short_name (e.g., '1', 'Blue Line') or route_long_name (e.g., 'Downtown Express')",
		},
		"same_name_and_description": {
			Description:    "Route short name and long name are identical. These should provide different levels of detail for passenger information.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles:  []string{"routes.txt"},
			AffectedFields: []string{"route_short_name", "route_long_name"},
			Impact:         "Reduced information value for passengers",
			ExampleFix:     "Use route_short_name for '1' or 'Blue', route_long_name for 'Downtown Express'",
		},
		"duplicate_route_name": {
			Description:   "Multiple routes have the same name. This may cause confusion for passengers and should be differentiated.",
			GTFSReference: "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles: []string{"routes.txt"},
			Impact:        "Passenger confusion when identifying routes",
			ExampleFix:    "Ensure each route has a unique name combination of short_name and long_name",
		},
		"poor_color_contrast": {
			Description:    "Route colors have insufficient contrast for accessibility compliance. This affects colorblind users and accessibility standards.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#routestxt",
			AffectedFiles:  []string{"routes.txt"},
			AffectedFields: []string{"route_color", "route_text_color"},
			Impact:         "Accessibility compliance issues, poor user experience for colorblind users",
			ExampleFix:     "Use high contrast combinations like white text (#FFFFFF) on dark backgrounds (#000000) or vice versa",
		},
		"missing_stop_name": {
			Description:    "A required stop name is missing. Stop names are essential for passenger identification and navigation.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles:  []string{"stops.txt"},
			AffectedFields: []string{"stop_name"},
			Impact:         "Passengers cannot identify stops, accessibility issues",
			ExampleFix:     "Add descriptive stop names like 'Main St & 1st Ave' or 'Downtown Transit Center'",
		},
		"foreign_key_violation": {
			Description:   "A foreign key reference is invalid. The referenced record does not exist in the target file.",
			GTFSReference: "https://gtfs.org/schedule/reference/#field-definitions",
			Impact:        "Data integrity issues, broken relationships between files",
			ExampleFix:    "Ensure referenced IDs exist. For example, if trips.txt references route_id 'R1', ensure 'R1' exists in routes.txt",
		},
		"excessive_travel_speed": {
			Description:    "Travel speed between stops is unrealistically fast for the transport mode. This may indicate data errors or missing stops.",
			GTFSReference:  "https://gtfs.org/schedule/best-practices/#stop_timestxt",
			AffectedFiles:  []string{"stop_times.txt"},
			AffectedFields: []string{"arrival_time", "departure_time"},
			Impact:         "Unrealistic trip planning, passenger confusion",
			ExampleFix:     "Check for missing intermediate stops or correct travel times. Bus speeds should typically be under 100 km/h.",
		},
		"stop_name_missing_but_inherited": {
			Description: "Stop name is missing but can inherit from parent station. Consider adding explicit stop name for clarity.",
			Impact:      "Reduced clarity for passengers, dependency on parent station naming",
			ExampleFix:  "Add explicit stop_name even if it can inherit from parent_station",
		},
		"generic_stop_name": {
			Description: "Stop name is too generic (e.g., 'stop', 'station'). Use descriptive names to help passengers identify locations.",
			Impact:      "Poor passenger experience, difficulty identifying stops",
			ExampleFix:  "Replace 'Stop 1' with 'Main St & 1st Ave' or 'Downtown Transit Center'",
		},
		"invalid_bikes_allowed": {
			Description:    "Bikes allowed field contains an invalid value. Must be 0 (no info), 1 (bikes allowed), or 2 (bikes not allowed).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#tripstxt",
			AffectedFiles:  []string{"trips.txt"},
			AffectedFields: []string{"bikes_allowed"},
			Impact:         "Incorrect bike policy information for passengers",
			ExampleFix:     "Use 0 for no information, 1 if bikes are allowed, 2 if bikes are not allowed",
		},
		"attribution_without_role": {
			Description:    "Attribution has no role assigned. Each attribution must have at least one role (producer, operator, or authority).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#attributionstxt",
			AffectedFiles:  []string{"attributions.txt"},
			AffectedFields: []string{"is_producer", "is_operator", "is_authority"},
			Impact:         "Unclear attribution responsibilities, compliance issues",
			ExampleFix:     "Set at least one role to 1: is_producer=1, is_operator=0, is_authority=0",
		},

		// === RELATIONSHIP VALIDATION ERRORS ===
		"duplicate_stop_sequence": {
			Description:    "Duplicate stop sequence found in a trip. Each stop in a trip must have a unique sequence number.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stop_timestxt",
			AffectedFiles:  []string{"stop_times.txt"},
			AffectedFields: []string{"stop_sequence"},
			Impact:         "Trip routing errors, navigation issues",
			ExampleFix:     "Ensure stop_sequence values are unique within each trip: 1, 2, 3, 4... not 1, 2, 2, 3",
		},
		"decreasing_stop_sequence": {
			Description:    "Stop sequence decreases along a trip. Stop sequences should generally increase from start to end.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stop_timestxt",
			AffectedFiles:  []string{"stop_times.txt"},
			AffectedFields: []string{"stop_sequence"},
			Impact:         "Confusing trip progression, route planning issues",
			ExampleFix:     "Use increasing sequence numbers: 1, 2, 3, 4 instead of 4, 3, 2, 1",
		},
		"arrival_after_departure": {
			Description:    "Arrival time is after departure time at a stop. This creates impossible travel scenarios.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stop_timestxt",
			AffectedFiles:  []string{"stop_times.txt"},
			AffectedFields: []string{"arrival_time", "departure_time"},
			Impact:         "Impossible schedule, confuses trip planners",
			ExampleFix:     "Ensure arrival_time <= departure_time: arrival_time=14:30:00, departure_time=14:32:00",
		},

		// === BUSINESS LOGIC ERRORS ===
		"impossible_travel_time": {
			Description:    "Travel time between stops is impossible for the transport mode. This indicates data quality issues.",
			GTFSReference:  "https://gtfs.org/schedule/best-practices/#stop_timestxt",
			AffectedFiles:  []string{"stop_times.txt"},
			AffectedFields: []string{"arrival_time", "departure_time"},
			Impact:         "Unrealistic schedules, passenger confusion",
			ExampleFix:     "Check time calculations: ensure sufficient travel time between stops based on distance and mode",
		},
		"invalid_frequency_time_range": {
			Description:    "Frequency time range is invalid. Start time must be before end time for each frequency period.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#frequenciestxt",
			AffectedFiles:  []string{"frequencies.txt"},
			AffectedFields: []string{"start_time", "end_time"},
			Impact:         "Invalid service periods, schedule confusion",
			ExampleFix:     "Ensure start_time < end_time: start_time=06:00:00, end_time=22:00:00",
		},
		"invalid_headway": {
			Description:    "Frequency headway is invalid. Headway must be greater than 0 seconds.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#frequenciestxt",
			AffectedFiles:  []string{"frequencies.txt"},
			AffectedFields: []string{"headway_secs"},
			Impact:         "Service frequency errors, schedule planning issues",
			ExampleFix:     "Use positive headway values: headway_secs=900 (15 minutes) instead of 0 or negative",
		},
		"overlapping_frequency": {
			Description:    "Frequency periods overlap for the same trip. Each time period should be distinct and non-overlapping.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#frequenciestxt",
			AffectedFiles:  []string{"frequencies.txt"},
			AffectedFields: []string{"start_time", "end_time", "trip_id"},
			Impact:         "Schedule conflicts, operational confusion",
			ExampleFix:     "Ensure non-overlapping periods: Period 1: 06:00-12:00, Period 2: 12:00-18:00",
		},
		"invalid_transfer_type": {
			Description:    "Transfer type is invalid. Must be 0 (recommended), 1 (timed), 2 (minimum time), or 3 (not possible).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#transferstxt",
			AffectedFiles:  []string{"transfers.txt"},
			AffectedFields: []string{"transfer_type"},
			Impact:         "Incorrect transfer guidance for passengers",
			ExampleFix:     "Use valid values: 0=recommended, 1=timed, 2=minimum time required, 3=not possible",
		},
		"expired_feed": {
			Description:    "The feed has expired. Feeds should be updated regularly to provide current service information.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#feed_infotxt",
			AffectedFiles:  []string{"feed_info.txt"},
			AffectedFields: []string{"feed_end_date"},
			Impact:         "Feed will be rejected by trip planners, outdated service information",
			ExampleFix:     "Update feed_end_date to a current date: feed_end_date=20241231",
		},
		"insufficient_service_coverage": {
			Description:   "Service coverage is insufficient for the next period. Ensure adequate service is available for passengers.",
			GTFSReference: "https://gtfs.org/schedule/reference/#calendartxt",
			AffectedFiles: []string{"calendar.txt", "calendar_dates.txt"},
			Impact:        "Limited service availability, poor passenger experience",
			ExampleFix:    "Extend service dates or add more active service periods in calendar.txt",
		},

		// === ACCESSIBILITY ERRORS ===
		"invalid_pathway_mode": {
			Description:    "Pathway mode is invalid or not specified. Pathway modes must be valid GTFS pathway type codes (1-7).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#pathwaystxt",
			AffectedFiles:  []string{"pathways.txt"},
			AffectedFields: []string{"pathway_mode"},
			Impact:         "Accessibility navigation issues, compliance problems",
			ExampleFix:     "Use valid codes: 1=walkway, 2=stairs, 3=moving_sidewalk, 4=escalator, 5=elevator, 6=fare_gate, 7=exit_gate",
		},
		"unreasonable_level_index": {
			Description:    "Level index is outside reasonable bounds. Level indices should be between -50 and +50.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#levelstxt",
			AffectedFiles:  []string{"levels.txt"},
			AffectedFields: []string{"level_index"},
			Impact:         "Level navigation confusion, accessibility issues",
			ExampleFix:     "Use reasonable values: level_index=0 (ground), -1 (basement), 2 (second floor)",
		},

		// === FARE SYSTEM ERRORS ===
		"invalid_fare_price": {
			Description:    "Fare price is negative or has excessive precision. Prices should be non-negative with reasonable decimal places.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_attributestxt",
			AffectedFiles:  []string{"fare_attributes.txt"},
			AffectedFields: []string{"price"},
			Impact:         "Fare calculation errors, payment system issues",
			ExampleFix:     "Use valid prices: price=2.50 (not -2.50 or 2.123456)",
		},
		"invalid_payment_method": {
			Description:    "Payment method is invalid. Must be 0 (paid on board) or 1 (paid before boarding).",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_attributestxt",
			AffectedFiles:  []string{"fare_attributes.txt"},
			AffectedFields: []string{"payment_method"},
			Impact:         "Confusion about payment process for passengers",
			ExampleFix:     "Use 0 for pay-on-board or 1 for prepaid tickets/cards",
		},
		"empty_fare_rule": {
			Description:    "A fare rule has no qualifying conditions specified. At least one condition must be defined.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_rulestxt",
			AffectedFiles:  []string{"fare_rules.txt"},
			AffectedFields: []string{"route_id", "origin_id", "destination_id", "contains_id"},
			Impact:         "Ambiguous fare application, pricing confusion",
			ExampleFix:     "Specify at least one condition: route_id=R1 or origin_id=zone_A",
		},

		// === GEOGRAPHIC DATA ERRORS ===
		"suspicious_coordinate": {
			Description:    "Coordinates appear to be placeholder or error values (e.g., 0,0). This may indicate data import issues.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles:  []string{"stops.txt", "shapes.txt"},
			AffectedFields: []string{"stop_lat", "stop_lon", "shape_pt_lat", "shape_pt_lon"},
			Impact:         "Incorrect location display, navigation problems",
			ExampleFix:     "Replace with actual coordinates: stop_lat=40.748817, stop_lon=-73.985428",
		},
		"very_close_stops": {
			Description:    "Stops are located very close together (within 10 meters). This may indicate duplicate stops or data errors.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles:  []string{"stops.txt"},
			AffectedFields: []string{"stop_lat", "stop_lon"},
			Impact:         "Confusion for passengers, redundant data",
			ExampleFix:     "Review stops and consolidate duplicates or ensure accurate coordinates",
		},

		// === NETWORK TOPOLOGY ERRORS ===
		"isolated_stop": {
			Description:   "A stop cannot be reached by any trip. This creates disconnected network elements.",
			GTFSReference: "https://gtfs.org/schedule/reference/#stopstxt",
			AffectedFiles: []string{"stops.txt", "stop_times.txt"},
			Impact:        "Stop is unusable by passengers, waste of resources",
			ExampleFix:    "Add trips that serve this stop or remove if no longer needed",
		},

		// === FEED INFO ERRORS ===
		"invalid_feed_language": {
			Description:    "Feed language code is invalid. Must be a valid 2-letter ISO 639-1 language code.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#feed_infotxt",
			AffectedFiles:  []string{"feed_info.txt"},
			AffectedFields: []string{"feed_lang"},
			Impact:         "Language detection issues in transit apps",
			ExampleFix:     "Use valid codes: en, es, fr, de, ja, etc.",
		},

		// === TRIP AND SERVICE ERRORS ===
		"service_never_active": {
			Description:    "A service is defined but never active on any day. This creates unused service definitions.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#calendartxt",
			AffectedFiles:  []string{"calendar.txt"},
			AffectedFields: []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"},
			Impact:         "Data bloat, potential confusion",
			ExampleFix:     "Set at least one day to 1: monday=1, or remove unused service",
		},
		"unused_service": {
			Description:    "A service is defined but not used by any trips. This creates orphaned service definitions.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#tripstxt",
			AffectedFiles:  []string{"calendar.txt", "trips.txt"},
			AffectedFields: []string{"service_id"},
			Impact:         "Data bloat, maintenance overhead",
			ExampleFix:     "Remove unused services or add trips that reference them",
		},
		"invalid_currency_code": {
			Description:    "Currency code is invalid or not recognized. Must be a valid ISO 4217 3-letter currency code.",
			GTFSReference:  "https://gtfs.org/schedule/reference/#fare_attributestxt",
			AffectedFiles:  []string{"fare_attributes.txt"},
			AffectedFields: []string{"currency_type"},
			Impact:         "Fare calculation errors, currency display issues",
			ExampleFix:     "Use valid codes: USD, EUR, CAD, GBP, JPY, etc.",
		},
		"insufficient_coordinate_precision": {
			Description:    "Coordinate precision is insufficient (less than 4 decimal places). This affects location accuracy.",
			GTFSReference:  "https://gtfs.org/schedule/best-practices/#coordinate-precision",
			AffectedFiles:  []string{"stops.txt", "shapes.txt"},
			AffectedFields: []string{"stop_lat", "stop_lon", "shape_pt_lat", "shape_pt_lon"},
			Impact:         "Inaccurate location display, navigation errors",
			ExampleFix:     "Use at least 4 decimal places: 40.7488 instead of 40.75",
		},
		"validator_error": {
			Description: "A validator encountered an error during processing. This may indicate data corruption or validator issues.",
			Impact:      "Validation may be incomplete, some issues may be missed",
			ExampleFix:  "Check data file integrity and report issue if problem persists",
		},
	}

	if desc, exists := descriptions[code]; exists {
		return desc
	}

	// Generate a user-friendly description from the code name for unknown codes
	words := strings.Split(code, "_")
	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word)
	}
	return NoticeDescription{
		Description: strings.Join(words, " ") + ". This validation check identified an issue that should be reviewed and corrected.",
		Impact:      "Data quality issue that should be reviewed",
	}
}
