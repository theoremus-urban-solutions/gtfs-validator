package core

import "strings"

// The types the GTFS spec gives each field, as a table.
//
// Before this existed, each enum was checked by hand in the validator that
// cared about it, which meant the same rule written a dozen times and a dozen
// codes saying "this field holds a value the spec does not define". Canonical
// reports all of them as unexpected_enum_value with the field in the context,
// and the rest of the type errors as invalid_date, invalid_integer and so on.
// The table below is what makes that one implementation instead of a dozen.

// fieldType is how a field's value is interpreted.
type fieldType int

const (
	typeID fieldType = iota
	typeEnum
	typeDate
	typeTime
	typeInteger
	typeFloat
	typeCurrencyCode
	typeCurrencyAmount
	typePhone
)

// fieldSpec describes one field of one file.
type fieldSpec struct {
	Name string
	Type fieldType

	// Values lists the enum values the spec defines. Empty for non-enums.
	Values []int
	// ExtendedEnumMin and Max admit a contiguous block on top of Values, which
	// route_type needs for its extended types.
	ExtendedEnumMin int
	ExtendedEnumMax int

	// Min and Max bound a numeric field. HasMin and HasMax say whether each
	// applies, since zero is a meaningful bound.
	Min, Max       float64
	HasMin, HasMax bool

	// CurrencyField names the field holding the currency a typeCurrencyAmount
	// is denominated in.
	CurrencyField string
}

func enum(name string, values ...int) fieldSpec {
	return fieldSpec{Name: name, Type: typeEnum, Values: values}
}

func id(name string) fieldSpec {
	return fieldSpec{Name: name, Type: typeID}
}

func date(name string) fieldSpec {
	return fieldSpec{Name: name, Type: typeDate}
}

func timeOfDay(name string) fieldSpec {
	return fieldSpec{Name: name, Type: typeTime}
}

// nonNegativeInt is the common case: a count, a sequence or a duration, none
// of which can run backwards.
func nonNegativeInt(name string) fieldSpec {
	return fieldSpec{Name: name, Type: typeInteger, Min: 0, HasMin: true}
}

func positiveInt(name string) fieldSpec {
	return fieldSpec{Name: name, Type: typeInteger, Min: 1, HasMin: true}
}

func nonNegativeFloat(name string) fieldSpec {
	return fieldSpec{Name: name, Type: typeFloat, Min: 0, HasMin: true}
}

func boundedFloat(name string, min, max float64) fieldSpec {
	return fieldSpec{Name: name, Type: typeFloat, Min: min, Max: max, HasMin: true, HasMax: true}
}

// latitude and longitude are bounded floats with the ranges of the globe.
func latitude(name string) fieldSpec  { return boundedFloat(name, -90, 90) }
func longitude(name string) fieldSpec { return boundedFloat(name, -180, 180) }

// routeTypeSpec admits the basic types plus the extended block (100-1799),
// which the spec allows and many European feeds use.
func routeTypeSpec() fieldSpec {
	spec := enum("route_type", 0, 1, 2, 3, 4, 5, 6, 7, 11, 12)
	spec.ExtendedEnumMin = 100
	spec.ExtendedEnumMax = 1799
	return spec
}

// fieldSpecs maps each file to the fields whose type the spec constrains.
// Fields not listed here are free-form text, or are checked elsewhere:
// URLs, emails, timezones, language codes and colours by
// field_format_validator.go, which already reports them under canonical codes.
var fieldSpecs = map[string][]fieldSpec{
	"agency.txt": {
		id("agency_id"),
		{Name: "agency_phone", Type: typePhone},
	},
	StopsFile: {
		id("stop_id"),
		id("parent_station"),
		id("zone_id"),
		id("level_id"),
		latitude("stop_lat"),
		longitude("stop_lon"),
		enum("location_type", 0, 1, 2, 3, 4),
		enum("wheelchair_boarding", 0, 1, 2),
	},
	RoutesFile: {
		id("route_id"),
		id("agency_id"),
		routeTypeSpec(),
		enum("continuous_pickup", 0, 1, 2, 3),
		enum("continuous_drop_off", 0, 1, 2, 3),
		nonNegativeInt("route_sort_order"),
	},
	TripsFile: {
		id("trip_id"),
		id("route_id"),
		id("service_id"),
		id("shape_id"),
		id("block_id"),
		enum("direction_id", 0, 1),
		enum("wheelchair_accessible", 0, 1, 2),
		enum("bikes_allowed", 0, 1, 2),
	},
	"stop_times.txt": {
		id("trip_id"),
		id("stop_id"),
		nonNegativeInt("stop_sequence"),
		timeOfDay("arrival_time"),
		timeOfDay("departure_time"),
		timeOfDay("start_pickup_drop_off_window"),
		timeOfDay("end_pickup_drop_off_window"),
		enum("pickup_type", 0, 1, 2, 3),
		enum("drop_off_type", 0, 1, 2, 3),
		enum("continuous_pickup", 0, 1, 2, 3),
		enum("continuous_drop_off", 0, 1, 2, 3),
		enum("timepoint", 0, 1),
		nonNegativeFloat("shape_dist_traveled"),
	},
	CalendarFile: {
		id("service_id"),
		date("start_date"),
		date("end_date"),
		enum("monday", 0, 1),
		enum("tuesday", 0, 1),
		enum("wednesday", 0, 1),
		enum("thursday", 0, 1),
		enum("friday", 0, 1),
		enum("saturday", 0, 1),
		enum("sunday", 0, 1),
	},
	CalendarDatesFile: {
		id("service_id"),
		date("date"),
		enum("exception_type", 1, 2),
	},
	"fare_attributes.txt": {
		id("fare_id"),
		id("agency_id"),
		// The spec types price as a non-negative float, not as a currency
		// amount: "0.8" EUR is a well formed price and means eighty cents. Only
		// the Fares v2 amounts below are held to their currency's subunit.
		nonNegativeFloat("price"),
		{Name: "currency_type", Type: typeCurrencyCode},
		enum("payment_method", 0, 1),
		enum("transfers", 0, 1, 2),
		nonNegativeInt("transfer_duration"),
	},
	"fare_products.txt": {
		{Name: "amount", Type: typeCurrencyAmount, CurrencyField: "currency"},
		{Name: "currency", Type: typeCurrencyCode},
	},
	"fare_rules.txt": {
		id("fare_id"),
		id("route_id"),
		id("origin_id"),
		id("destination_id"),
		id("contains_id"),
	},
	"shapes.txt": {
		id("shape_id"),
		latitude("shape_pt_lat"),
		longitude("shape_pt_lon"),
		nonNegativeInt("shape_pt_sequence"),
		nonNegativeFloat("shape_dist_traveled"),
	},
	"frequencies.txt": {
		id("trip_id"),
		timeOfDay("start_time"),
		timeOfDay("end_time"),
		positiveInt("headway_secs"),
		enum("exact_times", 0, 1),
	},
	"transfers.txt": {
		id("from_stop_id"),
		id("to_stop_id"),
		id("from_route_id"),
		id("to_route_id"),
		id("from_trip_id"),
		id("to_trip_id"),
		enum("transfer_type", 0, 1, 2, 3, 4, 5),
		nonNegativeInt("min_transfer_time"),
	},
	"pathways.txt": {
		id("pathway_id"),
		id("from_stop_id"),
		id("to_stop_id"),
		enum("pathway_mode", 1, 2, 3, 4, 5, 6, 7),
		enum("is_bidirectional", 0, 1),
		nonNegativeFloat("length"),
		positiveInt("traversal_time"),
		nonNegativeInt("stair_count"),
		boundedFloat("max_slope", -1, 1),
		nonNegativeFloat("min_width"),
	},
	"levels.txt": {
		id("level_id"),
		{Name: "level_index", Type: typeFloat},
	},
	"feed_info.txt": {
		date("feed_start_date"),
		date("feed_end_date"),
		{Name: "feed_contact_email", Type: typeID},
	},
	"attributions.txt": {
		id("attribution_id"),
		id("agency_id"),
		id("route_id"),
		id("trip_id"),
		enum("is_producer", 0, 1),
		enum("is_operator", 0, 1),
		enum("is_authority", 0, 1),
	},
	"translations.txt": {
		id("record_id"),
		id("record_sub_id"),
	},
}

// rangeSpec is a pair of fields describing the two ends of one range.
type rangeSpec struct {
	Start, End string
	// EqualIsError says whether the two ends being the same value is a defect.
	// A calendar covering a single day is normal; a frequency window of zero
	// length generates no trips.
	EqualIsError bool
	EntityField  string
	// Kind selects how the two values are compared.
	Kind fieldType
}

// rangeSpecs lists the paired fields whose order the spec constrains.
var rangeSpecs = map[string][]rangeSpec{
	CalendarFile: {
		{Start: "start_date", End: "end_date", EntityField: "service_id", Kind: typeDate},
	},
	"feed_info.txt": {
		{Start: "feed_start_date", End: "feed_end_date", EntityField: "feed_publisher_name", Kind: typeDate},
	},
	"frequencies.txt": {
		{Start: "start_time", End: "end_time", EqualIsError: true, EntityField: "trip_id", Kind: typeTime},
	},
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

// isISO4217 reports whether a code is an ISO 4217 currency.
func isISO4217(code string) bool {
	return validCurrencyCodes[strings.ToUpper(code)]
}

// currencyMinorUnits lists the currencies whose subunit is not two decimal
// places. Everything absent from this table uses two, which is the common case.
var currencyMinorUnits = map[string]int{
	"BHD": 3, "BIF": 0, "CLP": 0, "DJF": 0, "GNF": 0, "IQD": 3, "ISK": 0,
	"JOD": 3, "JPY": 0, "KMF": 0, "KRW": 0, "KWD": 3, "LYD": 3, "OMR": 3,
	"PYG": 0, "RWF": 0, "TND": 3, "UGX": 0, "UYW": 4, "VND": 0, "VUV": 0,
	"XAF": 0, "XOF": 0, "XPF": 0,
}

// currencyDecimals returns how many decimal places an amount in a currency
// must have.
func currencyDecimals(code string) int {
	if decimals, listed := currencyMinorUnits[strings.ToUpper(code)]; listed {
		return decimals
	}
	return 2
}
