package relationship

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// AttributionValidator validates attribution definitions
type AttributionValidator struct{}

// NewAttributionValidator creates a new attribution validator
func NewAttributionValidator() *AttributionValidator {
	return &AttributionValidator{}
}

// AttributionInfo represents attribution information
type AttributionInfo struct {
	AttributionID    string
	AgencyID         string
	RouteID          string
	TripID           string
	OrganizationName string
	IsProducer       *bool
	IsOperator       *bool
	IsAuthority      *bool
	AttributionURL   string
	AttributionEmail string
	AttributionPhone string
	RowNumber        int
}

// Validate checks attribution definitions
func (v *AttributionValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	attributions := v.loadAttributions(loader)
	if len(attributions) == 0 {
		return // No attributions to validate
	}

	// Load reference data
	agencies := v.loadAgencyIDs(loader)
	routes := v.loadRouteIDs(loader)
	trips := v.loadTripIDs(loader)

	// Validate each attribution
	for _, attribution := range attributions {
		v.validateAttribution(container, attribution, agencies, routes, trips)
	}

	// Validate attribution uniqueness
	v.validateAttributionUniqueness(container, attributions)
}

// loadAttributions loads attribution information from attributions.txt
func (v *AttributionValidator) loadAttributions(loader *parser.FeedLoader) []*AttributionInfo {
	var attributions []*AttributionInfo

	reader, err := loader.GetFile("attributions.txt")
	if err != nil {
		return attributions
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "attributions.txt")
	if err != nil {
		return attributions
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		attribution := v.parseAttribution(row)
		if attribution != nil {
			attributions = append(attributions, attribution)
		}
	}

	return attributions
}

// parseAttribution parses an attribution record
func (v *AttributionValidator) parseAttribution(row *parser.CSVRow) *AttributionInfo {
	attribution := &AttributionInfo{
		RowNumber: row.RowNumber,
	}

	// Parse optional attribution_id
	if attributionID, hasAttributionID := row.Values["attribution_id"]; hasAttributionID {
		attribution.AttributionID = strings.TrimSpace(attributionID)
	}

	// Parse optional reference fields
	if agencyID, hasAgencyID := row.Values["agency_id"]; hasAgencyID {
		attribution.AgencyID = strings.TrimSpace(agencyID)
	}
	if routeID, hasRouteID := row.Values["route_id"]; hasRouteID {
		attribution.RouteID = strings.TrimSpace(routeID)
	}
	if tripID, hasTripID := row.Values["trip_id"]; hasTripID {
		attribution.TripID = strings.TrimSpace(tripID)
	}

	// Parse organization name (required)
	if orgName, hasOrgName := row.Values["organization_name"]; hasOrgName {
		attribution.OrganizationName = strings.TrimSpace(orgName)
	}

	// Parse boolean role fields
	if isProducerStr, hasIsProducer := row.Values["is_producer"]; hasIsProducer && strings.TrimSpace(isProducerStr) != "" {
		if isProducer, err := strconv.Atoi(strings.TrimSpace(isProducerStr)); err == nil {
			isProducerBool := isProducer == 1
			attribution.IsProducer = &isProducerBool
		}
	}
	if isOperatorStr, hasIsOperator := row.Values["is_operator"]; hasIsOperator && strings.TrimSpace(isOperatorStr) != "" {
		if isOperator, err := strconv.Atoi(strings.TrimSpace(isOperatorStr)); err == nil {
			isOperatorBool := isOperator == 1
			attribution.IsOperator = &isOperatorBool
		}
	}
	if isAuthorityStr, hasIsAuthority := row.Values["is_authority"]; hasIsAuthority && strings.TrimSpace(isAuthorityStr) != "" {
		if isAuthority, err := strconv.Atoi(strings.TrimSpace(isAuthorityStr)); err == nil {
			isAuthorityBool := isAuthority == 1
			attribution.IsAuthority = &isAuthorityBool
		}
	}

	// Parse contact fields
	if attributionURL, hasAttributionURL := row.Values["attribution_url"]; hasAttributionURL {
		attribution.AttributionURL = strings.TrimSpace(attributionURL)
	}
	if attributionEmail, hasAttributionEmail := row.Values["attribution_email"]; hasAttributionEmail {
		attribution.AttributionEmail = strings.TrimSpace(attributionEmail)
	}
	if attributionPhone, hasAttributionPhone := row.Values["attribution_phone"]; hasAttributionPhone {
		attribution.AttributionPhone = strings.TrimSpace(attributionPhone)
	}

	return attribution
}

// validateAttribution validates a single attribution record
func (v *AttributionValidator) validateAttribution(container *notice.NoticeContainer, attribution *AttributionInfo, agencies, routes, trips map[string]bool) {
	// Check that organization_name is provided
	if attribution.OrganizationName == "" {
		container.AddNotice(notice.NewMissingRequiredFieldNotice(
			"attributions.txt",
			"organization_name",
			attribution.RowNumber,
		))
	}

	// Check that at least one role is specified
	hasRole := (attribution.IsProducer != nil && *attribution.IsProducer) ||
		(attribution.IsOperator != nil && *attribution.IsOperator) ||
		(attribution.IsAuthority != nil && *attribution.IsAuthority)

	if !hasRole {
		container.AddNotice(notice.NewMissingAttributionRoleNotice(
			attribution.AttributionID,
			attribution.RowNumber,
		))
	}

	// Validate foreign key references
	if attribution.AgencyID != "" && !agencies[attribution.AgencyID] {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"attributions.txt",
			"agency_id",
			attribution.AgencyID,
			attribution.RowNumber,
			"agency.txt",
			"agency_id",
		))
	}

	if attribution.RouteID != "" && !routes[attribution.RouteID] {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"attributions.txt",
			"route_id",
			attribution.RouteID,
			attribution.RowNumber,
			"routes.txt",
			"route_id",
		))
	}

	if attribution.TripID != "" && !trips[attribution.TripID] {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"attributions.txt",
			"trip_id",
			attribution.TripID,
			attribution.RowNumber,
			"trips.txt",
			"trip_id",
		))
	}

	// Validate scope consistency
	v.validateAttributionScope(container, attribution)

	// Validate contact information format
	v.validateAttributionContact(container, attribution)
}

// validateAttributionScope validates the scope of attribution
func (v *AttributionValidator) validateAttributionScope(container *notice.NoticeContainer, attribution *AttributionInfo) {
	scopeCount := 0
	if attribution.AgencyID != "" {
		scopeCount++
	}
	if attribution.RouteID != "" {
		scopeCount++
	}
	if attribution.TripID != "" {
		scopeCount++
	}

	// More than one scope specified
	if scopeCount > 1 {
		container.AddNotice(notice.NewMultipleAttributionScopesNotice(
			attribution.AttributionID,
			attribution.RowNumber,
		))
	}

	// Check for conflicting scopes
	if attribution.TripID != "" && attribution.RouteID != "" {
		container.AddNotice(notice.NewConflictingAttributionScopeNotice(
			attribution.AttributionID,
			attribution.RowNumber,
		))
	}
}

// validateAttributionContact validates contact information
func (v *AttributionValidator) validateAttributionContact(container *notice.NoticeContainer, attribution *AttributionInfo) {
	// Validate URL format
	if attribution.AttributionURL != "" {
		if !strings.HasPrefix(attribution.AttributionURL, "http://") &&
			!strings.HasPrefix(attribution.AttributionURL, "https://") {
			container.AddNotice(notice.NewInvalidURLNotice(
				"attributions.txt",
				"attribution_url",
				attribution.AttributionURL,
				attribution.RowNumber,
			))
		}
	}

	// Validate email format (basic check)
	if attribution.AttributionEmail != "" {
		if !strings.Contains(attribution.AttributionEmail, "@") ||
			!strings.Contains(attribution.AttributionEmail, ".") {
			container.AddNotice(notice.NewInvalidEmailNotice(
				"attributions.txt",
				"attribution_email",
				attribution.AttributionEmail,
				attribution.RowNumber,
			))
		}
	}

	// Check if at least one contact method is provided
	hasContact := attribution.AttributionURL != "" ||
		attribution.AttributionEmail != "" ||
		attribution.AttributionPhone != ""

	if !hasContact {
		container.AddNotice(notice.NewMissingAttributionContactNotice(
			attribution.AttributionID,
			attribution.RowNumber,
		))
	}
}

// validateAttributionUniqueness checks for duplicate attributions
func (v *AttributionValidator) validateAttributionUniqueness(container *notice.NoticeContainer, attributions []*AttributionInfo) {
	// Check for duplicate attribution IDs
	if len(attributions) > 0 {
		attributionIDs := make(map[string]*AttributionInfo)
		for _, attribution := range attributions {
			if attribution.AttributionID != "" {
				if existing, exists := attributionIDs[attribution.AttributionID]; exists {
					container.AddNotice(notice.NewDuplicateKeyNotice(
						"attributions.txt",
						"attribution_id",
						attribution.AttributionID,
						attribution.RowNumber,
						existing.RowNumber,
					))
				} else {
					attributionIDs[attribution.AttributionID] = attribution
				}
			}
		}
	}

	// Check for duplicate scope attributions
	scopeMap := make(map[string]*AttributionInfo)
	for _, attribution := range attributions {
		var scopeKey string
		if attribution.AgencyID != "" {
			scopeKey = "agency:" + attribution.AgencyID
		} else if attribution.RouteID != "" {
			scopeKey = "route:" + attribution.RouteID
		} else if attribution.TripID != "" {
			scopeKey = "trip:" + attribution.TripID
		} else {
			scopeKey = "global"
		}

		if existing, exists := scopeMap[scopeKey]; exists {
			container.AddNotice(notice.NewDuplicateAttributionScopeNotice(
				attribution.AttributionID,
				existing.AttributionID,
				scopeKey,
				attribution.RowNumber,
				existing.RowNumber,
			))
		} else {
			scopeMap[scopeKey] = attribution
		}
	}
}

// loadAgencyIDs loads agency IDs for foreign key validation
func (v *AttributionValidator) loadAgencyIDs(loader *parser.FeedLoader) map[string]bool {
	agencies := make(map[string]bool)

	reader, err := loader.GetFile("agency.txt")
	if err != nil {
		return agencies
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "agency.txt")
	if err != nil {
		return agencies
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if agencyID, hasAgencyID := row.Values["agency_id"]; hasAgencyID {
			agencies[strings.TrimSpace(agencyID)] = true
		}
	}

	return agencies
}

// loadRouteIDs loads route IDs for foreign key validation
func (v *AttributionValidator) loadRouteIDs(loader *parser.FeedLoader) map[string]bool {
	routes := make(map[string]bool)

	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return routes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return routes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if routeID, hasRouteID := row.Values["route_id"]; hasRouteID {
			routes[strings.TrimSpace(routeID)] = true
		}
	}

	return routes
}

// loadTripIDs loads trip IDs for foreign key validation
func (v *AttributionValidator) loadTripIDs(loader *parser.FeedLoader) map[string]bool {
	trips := make(map[string]bool)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return trips
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return trips
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if tripID, hasTripID := row.Values["trip_id"]; hasTripID {
			trips[strings.TrimSpace(tripID)] = true
		}
	}

	return trips
}
