package entity

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// AttributionWithoutRoleValidator validates attribution has at least one role assigned
type AttributionWithoutRoleValidator struct{}

// NewAttributionWithoutRoleValidator creates a new attribution without role validator
func NewAttributionWithoutRoleValidator() *AttributionWithoutRoleValidator {
	return &AttributionWithoutRoleValidator{}
}

// AttributionRoleInfo represents attribution role information
type AttributionRoleInfo struct {
	AttributionID    string
	OrganizationName string
	IsProducer       bool
	IsOperator       bool
	IsAuthority      bool
	RowNumber        int
}

// Validate checks that attributions have at least one role assigned
func (v *AttributionWithoutRoleValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	attributions := v.loadAttributions(loader)

	for _, attribution := range attributions {
		v.validateAttributionRoles(container, attribution)
	}
}

// loadAttributions loads attribution information from attributions.txt
func (v *AttributionWithoutRoleValidator) loadAttributions(loader *parser.FeedLoader) []*AttributionRoleInfo {
	var attributions []*AttributionRoleInfo

	reader, err := loader.GetFile("attributions.txt")
	if err != nil {
		return attributions // File doesn't exist, no attributions to validate
	}
	defer reader.Close()

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
			continue
		}

		attribution := v.parseAttribution(row)
		if attribution != nil {
			attributions = append(attributions, attribution)
		}
	}

	return attributions
}

// parseAttribution parses attribution information
func (v *AttributionWithoutRoleValidator) parseAttribution(row *parser.CSVRow) *AttributionRoleInfo {
	attribution := &AttributionRoleInfo{
		RowNumber: row.RowNumber,
	}

	// Parse attribution_id (optional)
	if attributionID, hasID := row.Values["attribution_id"]; hasID {
		attribution.AttributionID = strings.TrimSpace(attributionID)
	}

	// Parse organization_name (required)
	if orgName, hasOrgName := row.Values["organization_name"]; hasOrgName {
		attribution.OrganizationName = strings.TrimSpace(orgName)
	} else {
		return nil // organization_name is required
	}

	// Parse role fields
	if producerStr, hasProducer := row.Values["is_producer"]; hasProducer && strings.TrimSpace(producerStr) != "" {
		if producer, err := strconv.Atoi(strings.TrimSpace(producerStr)); err == nil && producer == 1 {
			attribution.IsProducer = true
		}
	}

	if operatorStr, hasOperator := row.Values["is_operator"]; hasOperator && strings.TrimSpace(operatorStr) != "" {
		if operator, err := strconv.Atoi(strings.TrimSpace(operatorStr)); err == nil && operator == 1 {
			attribution.IsOperator = true
		}
	}

	if authorityStr, hasAuthority := row.Values["is_authority"]; hasAuthority && strings.TrimSpace(authorityStr) != "" {
		if authority, err := strconv.Atoi(strings.TrimSpace(authorityStr)); err == nil && authority == 1 {
			attribution.IsAuthority = true
		}
	}

	return attribution
}

// validateAttributionRoles validates that attribution has at least one role
func (v *AttributionWithoutRoleValidator) validateAttributionRoles(container *notice.NoticeContainer, attribution *AttributionRoleInfo) {
	hasAnyRole := attribution.IsProducer || attribution.IsOperator || attribution.IsAuthority

	if !hasAnyRole {
		container.AddNotice(notice.NewAttributionWithoutRoleNotice(
			attribution.AttributionID,
			attribution.OrganizationName,
			attribution.RowNumber,
		))
	}

	// Additional validation: check for suspicious combinations
	v.validateRoleCombinations(container, attribution)
}

// validateRoleCombinations validates role combinations for consistency
func (v *AttributionWithoutRoleValidator) validateRoleCombinations(container *notice.NoticeContainer, attribution *AttributionRoleInfo) {
	roleCount := 0
	if attribution.IsProducer {
		roleCount++
	}
	if attribution.IsOperator {
		roleCount++
	}
	if attribution.IsAuthority {
		roleCount++
	}

	// Info notice if organization has all three roles (might be worth reviewing)
	if roleCount == 3 {
		container.AddNotice(notice.NewAttributionAllRolesNotice(
			attribution.AttributionID,
			attribution.OrganizationName,
			attribution.RowNumber,
		))
	}

	// Info notice if organization name suggests a specific role but has different roles
	v.validateRoleConsistencyWithName(container, attribution)
}

// validateRoleConsistencyWithName checks if organization name matches assigned roles
func (v *AttributionWithoutRoleValidator) validateRoleConsistencyWithName(container *notice.NoticeContainer, attribution *AttributionRoleInfo) {
	orgNameLower := strings.ToLower(attribution.OrganizationName)

	// Keywords that suggest specific roles
	operatorKeywords := []string{"transport", "transit", "bus", "metro", "railway", "operator", "service"}
	authorityKeywords := []string{"authority", "department", "ministry", "government", "city", "county", "state"}
	producerKeywords := []string{"data", "systems", "technology", "software", "solutions", "consulting"}

	hasOperatorKeyword := v.containsAnyKeyword(orgNameLower, operatorKeywords)
	hasAuthorityKeyword := v.containsAnyKeyword(orgNameLower, authorityKeywords)
	hasProducerKeyword := v.containsAnyKeyword(orgNameLower, producerKeywords)

	// Only report if there's a clear mismatch
	if hasOperatorKeyword && !attribution.IsOperator && (attribution.IsAuthority || attribution.IsProducer) {
		container.AddNotice(notice.NewAttributionRoleNameMismatchNotice(
			attribution.AttributionID,
			attribution.OrganizationName,
			"operator",
			attribution.RowNumber,
		))
	} else if hasAuthorityKeyword && !attribution.IsAuthority && (attribution.IsOperator || attribution.IsProducer) {
		container.AddNotice(notice.NewAttributionRoleNameMismatchNotice(
			attribution.AttributionID,
			attribution.OrganizationName,
			"authority",
			attribution.RowNumber,
		))
	} else if hasProducerKeyword && !attribution.IsProducer && (attribution.IsOperator || attribution.IsAuthority) {
		container.AddNotice(notice.NewAttributionRoleNameMismatchNotice(
			attribution.AttributionID,
			attribution.OrganizationName,
			"producer",
			attribution.RowNumber,
		))
	}
}

// containsAnyKeyword checks if text contains any of the given keywords
func (v *AttributionWithoutRoleValidator) containsAnyKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
