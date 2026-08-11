package core

import "strings"

// Phone numbers are checked for *possible length*, not validity — the same
// question the canonical validator asks. A number is rejected only when its
// national significant number cannot be that length in the feed's country: a
// bogus area code, an unassigned prefix or a disconnected line all pass, because
// nothing in a GTFS file can tell you otherwise and a validator that guessed
// would fail working feeds.
//
// The length is what catches the real defect, which is a truncated or mistyped
// number rather than a fictitious one.

// phoneLengths gives the national significant number lengths each country
// admits — that is, after the country calling code and any trunk prefix are
// removed. Countries absent from the table fall back to the E.164 bounds, which
// accept anything dialable anywhere; being permissive there is deliberate, since
// inventing a numbering plan would reject working feeds.
var phoneLengths = map[string]struct {
	CallingCode string
	Lengths     []int
}{
	"US": {"1", []int{10}},
	"CA": {"1", []int{10}},
	"GB": {"44", []int{9, 10}},
	"IE": {"353", []int{7, 8, 9}},
	"BG": {"359", []int{8, 9}},
	"DE": {"49", []int{6, 7, 8, 9, 10, 11}},
	"FR": {"33", []int{9}},
	"ES": {"34", []int{9}},
	"IT": {"39", []int{6, 7, 8, 9, 10, 11}},
	"NL": {"31", []int{9}},
	"BE": {"32", []int{8, 9}},
	"PT": {"351", []int{9}},
	"PL": {"48", []int{9}},
	"CZ": {"420", []int{9}},
	"AT": {"43", []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13}},
	"CH": {"41", []int{9}},
	"SE": {"46", []int{7, 8, 9}},
	"NO": {"47", []int{8}},
	"DK": {"45", []int{8}},
	"FI": {"358", []int{5, 6, 7, 8, 9, 10}},
	"GR": {"30", []int{10}},
	"RO": {"40", []int{9}},
	"HU": {"36", []int{8, 9}},
	"AU": {"61", []int{9}},
	"NZ": {"64", []int{8, 9, 10}},
	"BR": {"55", []int{10, 11}},
	"MX": {"52", []int{10}},
	"JP": {"81", []int{9, 10}},
	"IN": {"91", []int{10}},
}

// e164MinDigits and e164MaxDigits bound a subscriber number worldwide.
const (
	e164MinDigits = 4
	e164MaxDigits = 15
)

// keypadDigit maps a letter to the digit it shares a key with, so vanity
// numbers such as 1-800-FLOWERS are measured at the length they dial.
func keypadDigit(r rune) (rune, bool) {
	switch {
	case r >= 'a' && r <= 'z':
		r -= 'a' - 'A'
	case r < 'A' || r > 'Z':
		return 0, false
	}
	switch {
	case r <= 'C':
		return '2', true
	case r <= 'F':
		return '3', true
	case r <= 'I':
		return '4', true
	case r <= 'L':
		return '5', true
	case r <= 'O':
		return '6', true
	case r <= 'S':
		return '7', true
	case r <= 'V':
		return '8', true
	default:
		return '9', true
	}
}

// isPossiblePhoneNumber reports whether the value could be a phone number in
// the given country. An unrecognised or empty country code falls back to the
// E.164 bounds.
func isPossiblePhoneNumber(value string, countryCode string) bool {
	// An extension is dialled after the call connects and is not part of the
	// number's length.
	if cut := strings.IndexAny(value, "xX"); cut >= 0 {
		// Only when it separates digits, so it does not truncate a vanity
		// number that merely contains the letter.
		if strings.ContainsAny(value[:cut], "0123456789") {
			value = value[:cut]
		}
	}

	international := false
	var digits strings.Builder
	for i, r := range value {
		switch {
		case r >= '0' && r <= '9':
			digits.WriteRune(r)
		case r == '+':
			if i != 0 {
				return false // a plus anywhere but the front is not a number
			}
			international = true
		case strings.ContainsRune("-()., /", r):
			// Separators phone numbers are conventionally written with.
		default:
			if mapped, ok := keypadDigit(r); ok {
				digits.WriteRune(mapped)
				continue
			}
			return false
		}
	}

	national := digits.String()
	if national == "" {
		return false
	}

	plan, known := phoneLengths[strings.ToUpper(countryCode)]
	if !known {
		return len(national) >= e164MinDigits && len(national) <= e164MaxDigits
	}

	// Strip the country calling code when it is written out, then the trunk
	// prefix, leaving the national significant number the plan describes.
	if international || strings.HasPrefix(national, plan.CallingCode) {
		if trimmed := strings.TrimPrefix(national, plan.CallingCode); trimmed != national {
			for _, length := range plan.Lengths {
				if len(trimmed) == length {
					return true
				}
			}
		}
	}
	national = strings.TrimPrefix(national, "0")

	for _, length := range plan.Lengths {
		if len(national) == length {
			return true
		}
	}
	return false
}
