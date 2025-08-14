package fare

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// FareValidator validates fare system definitions
type FareValidator struct{}

// NewFareValidator creates a new fare validator
func NewFareValidator() *FareValidator {
	return &FareValidator{}
}

// validPaymentMethods contains valid GTFS payment methods
var validPaymentMethods = map[int]bool{
	0: true, // Fare is paid on board
	1: true, // Fare must be paid before boarding
}

// validTransfers contains valid GTFS transfer values
var validTransfers = map[int]bool{
	0: true, // No transfers permitted
	1: true, // Passengers may transfer once
	2: true, // Passengers may transfer twice
}

// FareAttributeInfo represents fare attribute information
type FareAttributeInfo struct {
	FareID           string
	Price            string
	CurrencyType     string
	PaymentMethod    *int
	Transfers        *int
	AgencyID         string
	TransferDuration *int
	RowNumber        int
}

// FareRuleInfo represents fare rule information
type FareRuleInfo struct {
	FareID        string
	RouteID       string
	OriginID      string
	DestinationID string
	ContainsID    string
	RowNumber     int
}

// Validate checks fare system definitions
func (v *FareValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load fare attributes
	fareAttributes := v.loadFareAttributes(loader)

	// Load fare rules
	fareRules := v.loadFareRules(loader)

	// Validate fare attributes
	for _, fareAttr := range fareAttributes {
		v.validateFareAttribute(container, fareAttr)
	}

	// Validate fare rules
	for _, fareRule := range fareRules {
		v.validateFareRule(container, fareRule, fareAttributes)
	}

	// Check for unused fare attributes
	v.validateUnusedFareAttributes(container, fareAttributes, fareRules)
}

// loadFareAttributes loads fare attributes from fare_attributes.txt
func (v *FareValidator) loadFareAttributes(loader *parser.FeedLoader) map[string]*FareAttributeInfo {
	fareAttributes := make(map[string]*FareAttributeInfo)

	reader, err := loader.GetFile("fare_attributes.txt")
	if err != nil {
		return fareAttributes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "fare_attributes.txt")
	if err != nil {
		return fareAttributes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		fareAttr := v.parseFareAttribute(row)
		if fareAttr != nil {
			fareAttributes[fareAttr.FareID] = fareAttr
		}
	}

	return fareAttributes
}

// parseFareAttribute parses a fare attribute record
func (v *FareValidator) parseFareAttribute(row *parser.CSVRow) *FareAttributeInfo {
	fareID, hasFareID := row.Values["fare_id"]
	price, hasPrice := row.Values["price"]
	currencyType, hasCurrencyType := row.Values["currency_type"]

	if !hasFareID || !hasPrice || !hasCurrencyType {
		return nil
	}

	fareAttr := &FareAttributeInfo{
		FareID:       strings.TrimSpace(fareID),
		Price:        strings.TrimSpace(price),
		CurrencyType: strings.TrimSpace(currencyType),
		RowNumber:    row.RowNumber,
	}

	// Parse optional fields
	if paymentMethodStr, hasPaymentMethod := row.Values["payment_method"]; hasPaymentMethod && strings.TrimSpace(paymentMethodStr) != "" {
		if paymentMethod, err := strconv.Atoi(strings.TrimSpace(paymentMethodStr)); err == nil {
			fareAttr.PaymentMethod = &paymentMethod
		}
	}

	if transfersStr, hasTransfers := row.Values["transfers"]; hasTransfers && strings.TrimSpace(transfersStr) != "" {
		if transfers, err := strconv.Atoi(strings.TrimSpace(transfersStr)); err == nil {
			fareAttr.Transfers = &transfers
		}
	}

	if agencyID, hasAgencyID := row.Values["agency_id"]; hasAgencyID {
		fareAttr.AgencyID = strings.TrimSpace(agencyID)
	}

	if transferDurationStr, hasTransferDuration := row.Values["transfer_duration"]; hasTransferDuration && strings.TrimSpace(transferDurationStr) != "" {
		if transferDuration, err := strconv.Atoi(strings.TrimSpace(transferDurationStr)); err == nil {
			fareAttr.TransferDuration = &transferDuration
		}
	}

	return fareAttr
}

// loadFareRules loads fare rules from fare_rules.txt
func (v *FareValidator) loadFareRules(loader *parser.FeedLoader) []*FareRuleInfo {
	var fareRules []*FareRuleInfo

	reader, err := loader.GetFile("fare_rules.txt")
	if err != nil {
		return fareRules
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "fare_rules.txt")
	if err != nil {
		return fareRules
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		fareRule := v.parseFareRule(row)
		if fareRule != nil {
			fareRules = append(fareRules, fareRule)
		}
	}

	return fareRules
}

// parseFareRule parses a fare rule record
func (v *FareValidator) parseFareRule(row *parser.CSVRow) *FareRuleInfo {
	fareID, hasFareID := row.Values["fare_id"]
	if !hasFareID {
		return nil
	}

	fareRule := &FareRuleInfo{
		FareID:    strings.TrimSpace(fareID),
		RowNumber: row.RowNumber,
	}

	// Parse optional fields
	if routeID, hasRouteID := row.Values["route_id"]; hasRouteID {
		fareRule.RouteID = strings.TrimSpace(routeID)
	}

	if originID, hasOriginID := row.Values["origin_id"]; hasOriginID {
		fareRule.OriginID = strings.TrimSpace(originID)
	}

	if destinationID, hasDestinationID := row.Values["destination_id"]; hasDestinationID {
		fareRule.DestinationID = strings.TrimSpace(destinationID)
	}

	if containsID, hasContainsID := row.Values["contains_id"]; hasContainsID {
		fareRule.ContainsID = strings.TrimSpace(containsID)
	}

	return fareRule
}

// validateFareAttribute validates a single fare attribute
func (v *FareValidator) validateFareAttribute(container *notice.NoticeContainer, fareAttr *FareAttributeInfo) {
	// Validate price format
	v.validatePrice(container, fareAttr)

	// Validate payment method
	if fareAttr.PaymentMethod != nil && !validPaymentMethods[*fareAttr.PaymentMethod] {
		container.AddNotice(notice.NewInvalidPaymentMethodNotice(
			fareAttr.FareID,
			*fareAttr.PaymentMethod,
			fareAttr.RowNumber,
		))
	}

	// Validate transfers
	if fareAttr.Transfers != nil {
		if *fareAttr.Transfers < 0 {
			container.AddNotice(notice.NewInvalidTransfersNotice(
				fareAttr.FareID,
				*fareAttr.Transfers,
				fareAttr.RowNumber,
			))
		} else if *fareAttr.Transfers > 2 && !validTransfers[*fareAttr.Transfers] {
			// Allow unlimited transfers (any value > 2)
			if *fareAttr.Transfers == 3 || *fareAttr.Transfers > 10 {
				container.AddNotice(notice.NewUnusualTransferValueNotice(
					fareAttr.FareID,
					*fareAttr.Transfers,
					fareAttr.RowNumber,
				))
			}
		}
	}

	// Validate transfer duration
	if fareAttr.TransferDuration != nil {
		if *fareAttr.TransferDuration < 0 {
			container.AddNotice(notice.NewInvalidTransferDurationNotice(
				fareAttr.FareID,
				*fareAttr.TransferDuration,
				fareAttr.RowNumber,
			))
		}

		// Check if transfer duration is provided but transfers is 0
		if fareAttr.Transfers != nil && *fareAttr.Transfers == 0 {
			container.AddNotice(notice.NewUnnecessaryTransferDurationNotice(
				fareAttr.FareID,
				*fareAttr.TransferDuration,
				fareAttr.RowNumber,
			))
		}
	}
}

// validatePrice validates the price field format
func (v *FareValidator) validatePrice(container *notice.NoticeContainer, fareAttr *FareAttributeInfo) {
	if fareAttr.Price == "" {
		return // Other validators handle missing price
	}

	// Try to parse as float
	price, err := strconv.ParseFloat(fareAttr.Price, 64)
	if err != nil {
		container.AddNotice(notice.NewInvalidFarePriceNotice(
			fareAttr.FareID,
			fareAttr.Price,
			fareAttr.RowNumber,
			"Price must be a valid number",
		))
		return
	}

	// Price should not be negative
	if price < 0 {
		container.AddNotice(notice.NewInvalidFarePriceNotice(
			fareAttr.FareID,
			fareAttr.Price,
			fareAttr.RowNumber,
			"Price cannot be negative",
		))
	}

	// Check for excessive decimal places (more than 4)
	priceStr := fareAttr.Price
	if dotIndex := strings.Index(priceStr, "."); dotIndex != -1 {
		decimals := len(priceStr) - dotIndex - 1
		if decimals > 4 {
			container.AddNotice(notice.NewExcessivePricePrecisionNotice(
				fareAttr.FareID,
				fareAttr.Price,
				decimals,
				fareAttr.RowNumber,
			))
		}
	}
}

// validateFareRule validates a single fare rule
func (v *FareValidator) validateFareRule(container *notice.NoticeContainer, fareRule *FareRuleInfo, fareAttributes map[string]*FareAttributeInfo) {
	// Validate fare_id reference
	if _, exists := fareAttributes[fareRule.FareID]; !exists {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"fare_rules.txt",
			"fare_id",
			fareRule.FareID,
			fareRule.RowNumber,
			"fare_attributes.txt",
			"fare_id",
		))
	}

	// Validate that at least one rule field is specified
	hasRule := fareRule.RouteID != "" || fareRule.OriginID != "" ||
		fareRule.DestinationID != "" || fareRule.ContainsID != ""

	if !hasRule {
		container.AddNotice(notice.NewEmptyFareRuleNotice(
			fareRule.FareID,
			fareRule.RowNumber,
		))
	}

	// Validate logical combinations
	v.validateFareRuleLogic(container, fareRule)
}

// validateFareRuleLogic validates logical consistency of fare rules
func (v *FareValidator) validateFareRuleLogic(container *notice.NoticeContainer, fareRule *FareRuleInfo) {
	// If origin_id and destination_id are the same
	if fareRule.OriginID != "" && fareRule.DestinationID != "" &&
		fareRule.OriginID == fareRule.DestinationID {
		container.AddNotice(notice.NewSameOriginDestinationNotice(
			fareRule.FareID,
			fareRule.OriginID,
			fareRule.RowNumber,
		))
	}

	// Contains_id should not be used with origin/destination
	if fareRule.ContainsID != "" && (fareRule.OriginID != "" || fareRule.DestinationID != "") {
		container.AddNotice(notice.NewConflictingFareRuleFieldsNotice(
			fareRule.FareID,
			fareRule.RowNumber,
		))
	}
}

// validateUnusedFareAttributes checks for fare attributes that are never used
func (v *FareValidator) validateUnusedFareAttributes(container *notice.NoticeContainer, fareAttributes map[string]*FareAttributeInfo, fareRules []*FareRuleInfo) {
	// Create set of used fare IDs
	usedFareIDs := make(map[string]bool)
	for _, fareRule := range fareRules {
		usedFareIDs[fareRule.FareID] = true
	}

	// Check for unused fare attributes
	for fareID, fareAttr := range fareAttributes {
		if !usedFareIDs[fareID] {
			container.AddNotice(notice.NewUnusedFareAttributeNotice(
				fareID,
				fareAttr.RowNumber,
			))
		}
	}
}
