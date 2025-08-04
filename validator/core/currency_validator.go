package core

import (
	"io"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// CurrencyValidator validates currency codes in fare_attributes.txt
type CurrencyValidator struct{}

// NewCurrencyValidator creates a new currency validator
func NewCurrencyValidator() *CurrencyValidator {
	return &CurrencyValidator{}
}

// validCurrencyCodes contains ISO 4217 currency codes
var validCurrencyCodes = map[string]bool{
	"AED": true, "AFN": true, "ALL": true, "AMD": true, "ANG": true, "AOA": true, "ARS": true, "AUD": true,
	"AWG": true, "AZN": true, "BAM": true, "BBD": true, "BDT": true, "BGN": true, "BHD": true, "BIF": true,
	"BMD": true, "BND": true, "BOB": true, "BRL": true, "BSD": true, "BTN": true, "BWP": true, "BYN": true,
	"BZD": true, "CAD": true, "CDF": true, "CHF": true, "CLP": true, "CNY": true, "COP": true, "CRC": true,
	"CUC": true, "CUP": true, "CVE": true, "CZK": true, "DJF": true, "DKK": true, "DOP": true, "DZD": true,
	"EGP": true, "ERN": true, "ETB": true, "EUR": true, "FJD": true, "FKP": true, "GBP": true, "GEL": true,
	"GGP": true, "GHS": true, "GIP": true, "GMD": true, "GNF": true, "GTQ": true, "GYD": true, "HKD": true,
	"HNL": true, "HRK": true, "HTG": true, "HUF": true, "IDR": true, "ILS": true, "IMP": true, "INR": true,
	"IQD": true, "IRR": true, "ISK": true, "JEP": true, "JMD": true, "JOD": true, "JPY": true, "KES": true,
	"KGS": true, "KHR": true, "KMF": true, "KPW": true, "KRW": true, "KWD": true, "KYD": true, "KZT": true,
	"LAK": true, "LBP": true, "LKR": true, "LRD": true, "LSL": true, "LYD": true, "MAD": true, "MDL": true,
	"MGA": true, "MKD": true, "MMK": true, "MNT": true, "MOP": true, "MRU": true, "MUR": true, "MVR": true,
	"MWK": true, "MXN": true, "MYR": true, "MZN": true, "NAD": true, "NGN": true, "NIO": true, "NOK": true,
	"NPR": true, "NZD": true, "OMR": true, "PAB": true, "PEN": true, "PGK": true, "PHP": true, "PKR": true,
	"PLN": true, "PYG": true, "QAR": true, "RON": true, "RSD": true, "RUB": true, "RWF": true, "SAR": true,
	"SBD": true, "SCR": true, "SDG": true, "SEK": true, "SGD": true, "SHP": true, "SLE": true, "SLL": true,
	"SOS": true, "SRD": true, "STN": true, "SYP": true, "SZL": true, "THB": true, "TJS": true, "TMT": true,
	"TND": true, "TOP": true, "TRY": true, "TTD": true, "TVD": true, "TWD": true, "TZS": true, "UAH": true,
	"UGX": true, "USD": true, "UYU": true, "UZS": true, "VED": true, "VES": true, "VND": true, "VUV": true,
	"WST": true, "XAF": true, "XCD": true, "XDR": true, "XOF": true, "XPF": true, "YER": true, "ZAR": true,
	"ZMW": true, "ZWL": true,
}

// Validate checks currency codes in fare_attributes.txt
func (v *CurrencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	v.validateFileCurrency(loader, container, "fare_attributes.txt", "currency_type")
}

// validateFileCurrency validates currency field in a specific file
func (v *CurrencyValidator) validateFileCurrency(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, fieldName string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, filename)
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

		if value, exists := row.Values[fieldName]; exists && strings.TrimSpace(value) != "" {
			v.validateCurrencyCode(container, filename, fieldName, strings.TrimSpace(value), row.RowNumber)
		}
	}
}

// validateCurrencyCode validates a single currency code
func (v *CurrencyValidator) validateCurrencyCode(container *notice.NoticeContainer, filename string, fieldName string, currencyCode string, rowNumber int) {
	// Currency codes should be exactly 3 uppercase letters
	if len(currencyCode) != 3 {
		container.AddNotice(notice.NewInvalidCurrencyCodeNotice(
			filename,
			fieldName,
			currencyCode,
			rowNumber,
			"Currency code must be exactly 3 characters",
		))
		return
	}

	// Check if it's a valid ISO 4217 currency code
	upperCode := strings.ToUpper(currencyCode)
	if !validCurrencyCodes[upperCode] {
		container.AddNotice(notice.NewInvalidCurrencyCodeNotice(
			filename,
			fieldName,
			currencyCode,
			rowNumber,
			"Unknown ISO 4217 currency code",
		))
		return
	}

	// Check for proper case (should be uppercase)
	if currencyCode != upperCode {
		container.AddNotice(notice.NewInvalidCurrencyCodeNotice(
			filename,
			fieldName,
			currencyCode,
			rowNumber,
			"Currency code should be uppercase",
		))
	}
}
