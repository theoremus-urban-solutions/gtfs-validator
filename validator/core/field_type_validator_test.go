package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestFieldTypeValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "well typed feed",
			files: map[string]string{
				"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,location_type,wheelchair_boarding\n1,Main St,34.05,-118.25,0,1",
				"routes.txt":     "route_id,agency_id,route_short_name,route_type\n1,1,Red,3",
				"trips.txt":      "route_id,service_id,trip_id,direction_id,wheelchair_accessible,bikes_allowed\n1,S1,T1,0,1,1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Every field matches the type the spec gives it",
		},

		// Enums. All of these were separate codes before; canonical reports
		// them as one, with the field name in the context.
		{
			name: "location_type out of range",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,Main St,34.05,-118.25,5",
			},
			expectedNoticeCodes: []string{"unexpected_enum_value"},
			description:         "location_type only runs 0-4; 5 parses as an integer and is simply not one of them",
		},
		{
			name: "non-numeric enum value",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,Main St,34.05,-118.25,platform",
			},
			expectedNoticeCodes: []string{"invalid_integer"},
			description:         "An enum is an integer first: a non-numeric value has no range to be outside of, so it is a bad integer rather than an unexpected enum",
		},
		{
			name: "several bad enums in one row",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,direction_id,wheelchair_accessible,bikes_allowed\n1,S1,T1,2,3,3",
			},
			expectedNoticeCodes: []string{"unexpected_enum_value", "unexpected_enum_value", "unexpected_enum_value"},
			description:         "One notice per offending field",
		},
		{
			name: "extended route types are valid",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type\n1,1,Red,100\n2,1,Blue,1799",
			},
			expectedNoticeCodes: []string{},
			description:         "The extended block 100-1799 is part of the spec",
		},
		{
			name: "route type above the extended block",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type\n1,1,Red,1800",
			},
			expectedNoticeCodes: []string{"unexpected_enum_value"},
			description:         "1800 is past the end of the extended block",
		},
		{
			name: "calendar day flags",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,2,1,1,1,1,0,0,20250101,20251231",
			},
			expectedNoticeCodes: []string{"unexpected_enum_value"},
			description:         "A day flag is 0 or 1",
		},

		// Numbers.
		{
			name: "negative stop_sequence",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,-1",
			},
			expectedNoticeCodes: []string{"number_out_of_range"},
			description:         "A sequence cannot run backwards",
		},
		{
			name: "negative shape_dist_traveled",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,1,1,-100.5",
			},
			expectedNoticeCodes: []string{"number_out_of_range"},
			description:         "Distance along a shape is measured forwards",
		},
		{
			name: "latitude past the pole",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,91.0,-118.25",
			},
			expectedNoticeCodes: []string{"number_out_of_range"},
			description:         "Latitude runs -90 to 90",
		},
		{
			name: "longitude past the antimeridian",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,200.0",
			},
			expectedNoticeCodes: []string{"number_out_of_range"},
			description:         "Longitude runs -180 to 180",
		},
		{
			name: "non-numeric coordinate",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,north,-118.25",
			},
			expectedNoticeCodes: []string{"invalid_float"},
			description:         "A coordinate that does not parse is a float error",
		},
		{
			name: "non-numeric integer field",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,first",
			},
			expectedNoticeCodes: []string{"invalid_integer"},
			description:         "stop_sequence must parse as an integer",
		},
		{
			name: "headway of zero",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,06:00:00,22:00:00,0",
			},
			expectedNoticeCodes: []string{"number_out_of_range"},
			description:         "A headway of zero generates infinitely many trips",
		},

		// Dates and times.
		{
			name: "malformed date",
			files: map[string]string{
				CalendarDatesFile: "service_id,date,exception_type\nS1,2025-01-01,1",
			},
			expectedNoticeCodes: []string{"invalid_date"},
			description:         "Dates are YYYYMMDD with no separators",
		},
		{
			name: "malformed time",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,8am,08:00:00,1,1",
			},
			expectedNoticeCodes: []string{"invalid_time"},
			description:         "Times are H:MM:SS or longer",
		},
		{
			name: "hours past 24 are valid",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,25:10:00,25:10:00,1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "A trip continuing past midnight keeps counting hours",
		},

		// Ranges.
		{
			name: "calendar ends before it starts",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20251231,20250101",
			},
			expectedNoticeCodes: []string{"start_and_end_range_out_of_order"},
			description:         "end_date must not precede start_date",
		},
		{
			name: "calendar covering a single day",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20250101",
			},
			expectedNoticeCodes: []string{},
			description:         "A one-day service is normal",
		},
		{
			name: "frequency window of zero length",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,06:00:00,06:00:00,600",
			},
			expectedNoticeCodes: []string{"start_and_end_range_equal"},
			description:         "A window that starts and ends at once generates no trips",
		},

		// Currency.
		{
			name: "unknown currency code",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method\nF1,2.50,XYZ,0",
			},
			expectedNoticeCodes: []string{"invalid_currency"},
			description:         "XYZ is not ISO 4217",
		},
		{
			name: "fare price written to one decimal place",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method\nF1,0.8,EUR,0",
			},
			expectedNoticeCodes: []string{},
			description:         "price is a plain float; 0.8 EUR is eighty cents",
		},
		{
			name: "fare price with more decimals than the euro has",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method\nF1,1.1,EUR,0\nF2,1.855,EUR,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Still a float: the subunit does not bound it",
		},
		{
			name: "fare price that is not a number",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method\nF1,free,EUR,0",
			},
			expectedNoticeCodes: []string{"invalid_float"},
			description:         "A price still has to parse",
		},
		{
			name: "fare product amount with too few decimals for its currency",
			files: map[string]string{
				"fare_products.txt": "fare_product_id,amount,currency\nP1,2.5,USD",
			},
			expectedNoticeCodes: []string{"invalid_currency_amount"},
			description:         "The dollar has two decimal places",
		},
		{
			name: "fare product amount with too many decimals for its currency",
			files: map[string]string{
				"fare_products.txt": "fare_product_id,amount,currency\nP1,1.855,EUR",
			},
			expectedNoticeCodes: []string{"invalid_currency_amount"},
			description:         "The euro has two decimal places, not three",
		},
		{
			name: "fare product amount matching its currency",
			files: map[string]string{
				"fare_products.txt": "fare_product_id,amount,currency\nP1,2.50,USD",
			},
			expectedNoticeCodes: []string{},
			description:         "Two decimal places is what the dollar asks for",
		},
		{
			name: "yen amount with no decimals",
			files: map[string]string{
				"fare_products.txt": "fare_product_id,amount,currency\nP1,200,JPY",
			},
			expectedNoticeCodes: []string{},
			description:         "The yen has no subunit",
		},
		{
			name: "yen amount written with decimals",
			files: map[string]string{
				"fare_products.txt": "fare_product_id,amount,currency\nP1,200.00,JPY",
			},
			expectedNoticeCodes: []string{"invalid_currency_amount"},
			description:         "The yen has no subunit to spend",
		},

		// Structure.
		{
			name: "empty column name",
			files: map[string]string{
				"stops.txt": "stop_id,,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{"empty_column_name"},
			description:         "A column with no name cannot be referred to",
		},
		{
			name: "id with non-ASCII characters",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nстоп1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{"non_ascii_or_non_printable_char"},
			description:         "IDs travel through systems that may not carry them intact",
		},

		// Absence is the required-field validator's business, not this one's.
		{
			name: "empty optional fields",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,wheelchair_boarding\n1,Main St,34.05,-118.25,,",
			},
			expectedNoticeCodes: []string{},
			description:         "An empty optional field has no type to violate",
		},
		{
			name: "whitespace-only optional field",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,direction_id\n1,S1,T1,   ",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace is trimmed before the field is read",
		},
		{
			name: "boundary enum values",
			files: map[string]string{
				"stops.txt":           "stop_id,stop_name,stop_lat,stop_lon,location_type,wheelchair_boarding\n1,Main St,34.05,-118.25,4,2",
				"trips.txt":           "route_id,service_id,trip_id,direction_id,wheelchair_accessible,bikes_allowed\n1,S1,T1,1,2,2",
				"transfers.txt":       "from_stop_id,to_stop_id,transfer_type\n1,2,3",
				CalendarDatesFile:     "service_id,date,exception_type\nS1,20250101,1",
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0",
			},
			expectedNoticeCodes: []string{},
			description:         "The ends of each enum's range are valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewFieldTypeValidator()

			validator.Validate(loader, container, gtfsvalidator.Config{})

			assertNoticeCodes(t, container, tt.expectedNoticeCodes, tt.description)
		})
	}
}

func TestFieldTypeValidator_New(t *testing.T) {
	if NewFieldTypeValidator() == nil {
		t.Error("NewFieldTypeValidator() returned nil")
	}
}
