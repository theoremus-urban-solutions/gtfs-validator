package business

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TransferValidator validates transfer definitions
type TransferValidator struct{}

// NewTransferValidator creates a new transfer validator
func NewTransferValidator() *TransferValidator {
	return &TransferValidator{}
}

// validTransferTypes contains valid GTFS transfer types
var validTransferTypes = map[int]bool{
	0: true, // Recommended transfer point
	1: true, // Timed transfer point
	2: true, // Minimum time required to transfer
	3: true, // Transfers not possible
}

// TransferInfo represents transfer information
type TransferInfo struct {
	FromStopID      string
	ToStopID        string
	TransferType    int
	MinTransferTime *int
	RowNumber       int
}

// Validate checks transfer definitions
func (v *TransferValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	transfers := v.loadTransfers(loader)

	// Load stop information for validation
	stops := v.loadStopIDs(loader)

	// Validate each transfer
	for _, transfer := range transfers {
		v.validateTransfer(container, transfer, stops)
	}

	// Check for duplicate transfers
	v.validateDuplicateTransfers(container, transfers)
}

// loadTransfers loads transfer information from transfers.txt
func (v *TransferValidator) loadTransfers(loader *parser.FeedLoader) []*TransferInfo {
	var transfers []*TransferInfo

	reader, err := loader.GetFile("transfers.txt")
	if err != nil {
		return transfers
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "transfers.txt")
	if err != nil {
		return transfers
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		transfer := v.parseTransfer(row)
		if transfer != nil {
			transfers = append(transfers, transfer)
		}
	}

	return transfers
}

// parseTransfer parses a transfer record
func (v *TransferValidator) parseTransfer(row *parser.CSVRow) *TransferInfo {
	fromStopID, hasFromStopID := row.Values["from_stop_id"]
	toStopID, hasToStopID := row.Values["to_stop_id"]
	transferTypeStr, hasTransferType := row.Values["transfer_type"]

	if !hasFromStopID || !hasToStopID || !hasTransferType {
		return nil
	}

	transferType, err := strconv.Atoi(strings.TrimSpace(transferTypeStr))
	if err != nil {
		return nil
	}

	transfer := &TransferInfo{
		FromStopID:   strings.TrimSpace(fromStopID),
		ToStopID:     strings.TrimSpace(toStopID),
		TransferType: transferType,
		RowNumber:    row.RowNumber,
	}

	// Parse min_transfer_time if present
	if minTransferTimeStr, hasMinTransferTime := row.Values["min_transfer_time"]; hasMinTransferTime && strings.TrimSpace(minTransferTimeStr) != "" {
		if minTransferTime, err := strconv.Atoi(strings.TrimSpace(minTransferTimeStr)); err == nil {
			transfer.MinTransferTime = &minTransferTime
		}
	}

	return transfer
}

// loadStopIDs loads all stop IDs from stops.txt
func (v *TransferValidator) loadStopIDs(loader *parser.FeedLoader) map[string]bool {
	stopIDs := make(map[string]bool)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return stopIDs
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return stopIDs
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if stopID, hasStopID := row.Values["stop_id"]; hasStopID {
			stopIDs[strings.TrimSpace(stopID)] = true
		}
	}

	return stopIDs
}

// validateTransfer validates a single transfer record
func (v *TransferValidator) validateTransfer(container *notice.NoticeContainer, transfer *TransferInfo, stops map[string]bool) {
	// Validate transfer type
	if !validTransferTypes[transfer.TransferType] {
		container.AddNotice(notice.NewInvalidTransferTypeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			transfer.TransferType,
			transfer.RowNumber,
		))
	}

	// Validate stop references
	if !stops[transfer.FromStopID] {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"transfers.txt",
			"from_stop_id",
			transfer.FromStopID,
			transfer.RowNumber,
			"stops.txt",
			"stop_id",
		))
	}

	if !stops[transfer.ToStopID] {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"transfers.txt",
			"to_stop_id",
			transfer.ToStopID,
			transfer.RowNumber,
			"stops.txt",
			"stop_id",
		))
	}

	// Validate transfer from/to same stop
	if transfer.FromStopID == transfer.ToStopID {
		container.AddNotice(notice.NewTransferToSameStopNotice(
			transfer.FromStopID,
			transfer.RowNumber,
		))
	}

	// Validate min_transfer_time requirements
	v.validateMinTransferTime(container, transfer)
}

// validateMinTransferTime validates min_transfer_time field
func (v *TransferValidator) validateMinTransferTime(container *notice.NoticeContainer, transfer *TransferInfo) {
	// min_transfer_time is required for transfer_type = 2
	if transfer.TransferType == 2 && transfer.MinTransferTime == nil {
		container.AddNotice(notice.NewMissingMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			transfer.RowNumber,
		))
		return
	}

	// min_transfer_time should not be used for transfer_type = 3
	if transfer.TransferType == 3 && transfer.MinTransferTime != nil {
		container.AddNotice(notice.NewUnnecessaryMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			*transfer.MinTransferTime,
			transfer.RowNumber,
		))
	}

	// Validate min_transfer_time value
	if transfer.MinTransferTime != nil {
		if *transfer.MinTransferTime < 0 {
			container.AddNotice(notice.NewNegativeMinTransferTimeNotice(
				transfer.FromStopID,
				transfer.ToStopID,
				*transfer.MinTransferTime,
				transfer.RowNumber,
			))
		}

		// Check for unreasonably long transfer times (more than 1 hour)
		if *transfer.MinTransferTime > 3600 {
			container.AddNotice(notice.NewUnreasonableMinTransferTimeNotice(
				transfer.FromStopID,
				transfer.ToStopID,
				*transfer.MinTransferTime,
				transfer.RowNumber,
			))
		}
	}
}

// validateDuplicateTransfers checks for duplicate transfer definitions
func (v *TransferValidator) validateDuplicateTransfers(container *notice.NoticeContainer, transfers []*TransferInfo) {
	transferMap := make(map[string]*TransferInfo)

	for _, transfer := range transfers {
		key := transfer.FromStopID + "->" + transfer.ToStopID

		if existingTransfer, exists := transferMap[key]; exists {
			container.AddNotice(notice.NewDuplicateTransferNotice(
				transfer.FromStopID,
				transfer.ToStopID,
				transfer.RowNumber,
				existingTransfer.RowNumber,
			))
		} else {
			transferMap[key] = transfer
		}
	}
}
