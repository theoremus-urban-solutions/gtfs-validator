package schema

// FareAttribute represents fare information from fare_attributes.txt
type FareAttribute struct {
	FareID           string  `csv:"fare_id"`
	Price            float64 `csv:"price"`
	CurrencyType     string  `csv:"currency_type"`
	PaymentMethod    int     `csv:"payment_method"`
	Transfers        int     `csv:"transfers"`
	AgencyID         string  `csv:"agency_id"`
	TransferDuration int     `csv:"transfer_duration"`
}
