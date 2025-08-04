package entity

import (
	"io"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// AgencyConsistencyValidator validates agency references and consistency
type AgencyConsistencyValidator struct{}

// NewAgencyConsistencyValidator creates a new agency consistency validator
func NewAgencyConsistencyValidator() *AgencyConsistencyValidator {
	return &AgencyConsistencyValidator{}
}

// Validate checks agency consistency rules
func (v *AgencyConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load agencies
	agencies := v.loadAgencies(loader)
	if len(agencies) == 0 {
		return // No agencies to validate
	}

	// Check if agency_id is required
	v.validateAgencyIdRequirement(loader, container, agencies)

	// Check route agency references
	v.validateRouteAgencyReferences(loader, container, agencies)
}

// AgencyInfo represents agency information
type AgencyInfo struct {
	AgencyID   string
	AgencyName string
	RowNumber  int
}

// loadAgencies loads agency information from agency.txt
func (v *AgencyConsistencyValidator) loadAgencies(loader *parser.FeedLoader) map[string]*AgencyInfo {
	agencies := make(map[string]*AgencyInfo)

	reader, err := loader.GetFile("agency.txt")
	if err != nil {
		return agencies
	}
	defer reader.Close()

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

		agencyID, hasAgencyID := row.Values["agency_id"]
		agencyName, hasAgencyName := row.Values["agency_name"]

		// Use empty string as key if no agency_id provided
		key := ""
		if hasAgencyID {
			key = strings.TrimSpace(agencyID)
		}

		name := ""
		if hasAgencyName {
			name = strings.TrimSpace(agencyName)
		}

		agencies[key] = &AgencyInfo{
			AgencyID:   key,
			AgencyName: name,
			RowNumber:  row.RowNumber,
		}
	}

	return agencies
}

// validateAgencyIdRequirement checks if agency_id is required
func (v *AgencyConsistencyValidator) validateAgencyIdRequirement(loader *parser.FeedLoader, container *notice.NoticeContainer, agencies map[string]*AgencyInfo) {
	// If there are multiple agencies, agency_id is required
	if len(agencies) > 1 {
		for _, agency := range agencies {
			if agency.AgencyID == "" {
				container.AddNotice(notice.NewMissingAgencyIdNotice(
					agency.AgencyName,
					agency.RowNumber,
				))
			}
		}
	}
}

// validateRouteAgencyReferences checks that routes reference valid agencies
func (v *AgencyConsistencyValidator) validateRouteAgencyReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, agencies map[string]*AgencyInfo) {
	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return // No routes file
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		routeID, hasRouteID := row.Values["route_id"]
		agencyID, hasAgencyID := row.Values["agency_id"]

		if !hasRouteID {
			continue
		}

		// Determine which agency this route should reference
		expectedAgencyID := ""
		if hasAgencyID {
			expectedAgencyID = strings.TrimSpace(agencyID)
		}

		// If no agency_id specified in route, use the single agency if there's only one
		if expectedAgencyID == "" && len(agencies) == 1 {
			// This is valid - routes can omit agency_id if there's only one agency
			continue
		}

		// Check if referenced agency exists
		if _, exists := agencies[expectedAgencyID]; !exists {
			container.AddNotice(notice.NewInvalidAgencyReferenceNotice(
				strings.TrimSpace(routeID),
				expectedAgencyID,
				row.RowNumber,
			))
		}

		// If multiple agencies exist but route has no agency_id, that's an error
		if expectedAgencyID == "" && len(agencies) > 1 {
			container.AddNotice(notice.NewMissingRouteAgencyIdNotice(
				strings.TrimSpace(routeID),
				row.RowNumber,
			))
		}
	}
}
