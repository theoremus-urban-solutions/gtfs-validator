package schema

// Agency represents a transit agency from agency.txt
type Agency struct {
	AgencyID       string `csv:"agency_id"`
	AgencyName     string `csv:"agency_name"`
	AgencyURL      string `csv:"agency_url"`
	AgencyTimezone string `csv:"agency_timezone"`
	AgencyLang     string `csv:"agency_lang"`
	AgencyPhone    string `csv:"agency_phone"`
	AgencyEmail    string `csv:"agency_email"`
	AgencyFareURL  string `csv:"agency_fare_url"`
}
