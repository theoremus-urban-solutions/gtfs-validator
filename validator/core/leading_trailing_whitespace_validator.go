package core

import (
	"io"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// LeadingTrailingWhitespaceValidator checks for fields with leading or trailing whitespace
type LeadingTrailingWhitespaceValidator struct{}

// NewLeadingTrailingWhitespaceValidator creates a new whitespace validator
func NewLeadingTrailingWhitespaceValidator() *LeadingTrailingWhitespaceValidator {
	return &LeadingTrailingWhitespaceValidator{}
}

// Validate checks for leading and trailing whitespace in GTFS fields
func (v *LeadingTrailingWhitespaceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Get list of all GTFS files to check
	gtfsFiles := []string{
		"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt",
		"calendar.txt", "calendar_dates.txt", "fare_attributes.txt",
		"fare_rules.txt", "shapes.txt", "frequencies.txt", "transfers.txt",
		"pathways.txt", "levels.txt", "feed_info.txt", "attributions.txt",
	}

	for _, filename := range gtfsFiles {
		if loader.HasFile(filename) {
			v.validateFile(loader, container, filename)
		}
	}
}

// validateFile validates a specific GTFS file for whitespace issues
func (v *LeadingTrailingWhitespaceValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return
	}

	// Get fields that should be checked for whitespace
	significantFields := v.getSignificantFields(filename)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// Check each field for whitespace issues
		for fieldName, fieldValue := range row.Values {
			// Skip empty fields
			if fieldValue == "" {
				continue
			}

			// Check if this field should be validated
			if v.shouldValidateField(fieldName, significantFields) {
				v.validateFieldWhitespace(container, filename, fieldName, fieldValue, row.RowNumber)
			}
		}
	}
}

// validateFieldWhitespace checks a specific field for whitespace issues
func (v *LeadingTrailingWhitespaceValidator) validateFieldWhitespace(container *notice.NoticeContainer, filename, fieldName, fieldValue string, rowNumber int) {
	trimmed := strings.TrimSpace(fieldValue)

	// Check for fields that are only whitespace
	if trimmed == "" {
		container.AddNotice(notice.NewWhitespaceOnlyFieldNotice(
			filename,
			fieldName,
			rowNumber,
		))
	}

	// Check for leading whitespace
	if strings.HasPrefix(fieldValue, " ") || strings.HasPrefix(fieldValue, "\t") {
		container.AddNotice(notice.NewLeadingWhitespaceNotice(
			filename,
			fieldName,
			fieldValue,
			rowNumber,
		))
	}

	// Check for trailing whitespace
	if strings.HasSuffix(fieldValue, " ") || strings.HasSuffix(fieldValue, "\t") {
		container.AddNotice(notice.NewTrailingWhitespaceNotice(
			filename,
			fieldName,
			fieldValue,
			rowNumber,
		))
	}

	// Check for excessive internal whitespace (multiple consecutive spaces)
	if strings.Contains(fieldValue, "  ") {
		container.AddNotice(notice.NewExcessiveWhitespaceNotice(
			filename,
			fieldName,
			fieldValue,
			rowNumber,
		))
	}
}

// shouldValidateField determines if a field should be checked for whitespace
func (v *LeadingTrailingWhitespaceValidator) shouldValidateField(fieldName string, significantFields map[string]bool) bool {
	// If no specific fields defined, validate all text fields
	if len(significantFields) == 0 {
		return v.isTextField(fieldName)
	}

	// Check if field is in the significant fields list
	return significantFields[fieldName]
}

// isTextField determines if a field typically contains text data
func (v *LeadingTrailingWhitespaceValidator) isTextField(fieldName string) bool {
	// Numeric and coordinate fields don't need whitespace validation as much
	numericFields := map[string]bool{
		"stop_lat": true, "stop_lon": true, "route_type": true,
		"direction_id": true, "location_type": true, "wheelchair_boarding": true,
		"wheelchair_accessible": true, "bikes_allowed": true, "stop_sequence": true,
		"pickup_type": true, "drop_off_type": true, "shape_dist_traveled": true,
		"timepoint": true, "monday": true, "tuesday": true, "wednesday": true,
		"thursday": true, "friday": true, "saturday": true, "sunday": true,
		"exception_type": true, "payment_method": true, "transfers": true,
		"transfer_duration": true, "shape_pt_lat": true, "shape_pt_lon": true,
		"shape_pt_sequence": true, "headway_secs": true, "exact_times": true,
		"transfer_type": true, "min_transfer_time": true, "pathway_mode": true,
		"is_bidirectional": true, "length": true, "traversal_time": true,
		"stair_count": true, "max_slope": true, "min_width": true,
		"signposted_as": true, "reversed_signposted_as": true,
	}

	return !numericFields[fieldName]
}

// getSignificantFields returns fields that are particularly important for whitespace validation
func (v *LeadingTrailingWhitespaceValidator) getSignificantFields(filename string) map[string]bool {
	switch filename {
	case "agency.txt":
		return map[string]bool{
			"agency_id": true, "agency_name": true, "agency_url": true,
			"agency_timezone": true, "agency_lang": true, "agency_phone": true,
			"agency_fare_url": true, "agency_email": true,
		}
	case "stops.txt":
		return map[string]bool{
			"stop_id": true, "stop_code": true, "stop_name": true,
			"stop_desc": true, "zone_id": true, "stop_url": true,
			"parent_station": true, "stop_timezone": true, "level_id": true,
			"platform_code": true,
		}
	case "routes.txt":
		return map[string]bool{
			"route_id": true, "agency_id": true, "route_short_name": true,
			"route_long_name": true, "route_desc": true, "route_url": true,
			"route_color": true, "route_text_color": true, "route_sort_order": true,
		}
	case "trips.txt":
		return map[string]bool{
			"route_id": true, "service_id": true, "trip_id": true,
			"trip_headsign": true, "trip_short_name": true, "block_id": true,
			"shape_id": true,
		}
	case "stop_times.txt":
		return map[string]bool{
			"trip_id": true, "arrival_time": true, "departure_time": true,
			"stop_id": true, "stop_headsign": true,
		}
	case "calendar.txt":
		return map[string]bool{
			"service_id": true, "start_date": true, "end_date": true,
		}
	case "calendar_dates.txt":
		return map[string]bool{
			"service_id": true, "date": true,
		}
	case "fare_attributes.txt":
		return map[string]bool{
			"fare_id": true, "price": true, "currency_type": true,
			"agency_id": true,
		}
	case "fare_rules.txt":
		return map[string]bool{
			"fare_id": true, "route_id": true, "origin_id": true,
			"destination_id": true, "contains_id": true,
		}
	case "shapes.txt":
		return map[string]bool{
			"shape_id": true,
		}
	case "feed_info.txt":
		return map[string]bool{
			"feed_publisher_name": true, "feed_publisher_url": true,
			"feed_lang": true, "feed_start_date": true, "feed_end_date": true,
			"feed_version": true, "feed_contact_email": true, "feed_contact_url": true,
		}
	case "frequencies.txt":
		return map[string]bool{
			"trip_id": true, "start_time": true, "end_time": true,
		}
	case "transfers.txt":
		return map[string]bool{
			"from_stop_id": true, "to_stop_id": true,
		}
	case "pathways.txt":
		return map[string]bool{
			"pathway_id": true, "from_stop_id": true, "to_stop_id": true,
		}
	case "levels.txt":
		return map[string]bool{
			"level_id": true, "level_index": true, "level_name": true,
		}
	case "attributions.txt":
		return map[string]bool{
			"attribution_id": true, "agency_id": true, "route_id": true,
			"trip_id": true, "organization_name": true, "attribution_url": true,
			"attribution_email": true, "attribution_phone": true,
		}
	default:
		return map[string]bool{} // Validate all text fields
	}
}
