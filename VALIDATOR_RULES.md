# GTFS Validator Rules Documentation

This document provides comprehensive documentation of all GTFS validators and the rules they implement. The GTFS validator performs over 100 different validation checks across all GTFS files and features.

## Overview

The validator is organized into several categories, each focusing on different aspects of GTFS feed validation:

- **[Core Validators](#core-validators)**: Basic file structure, format, and required fields
- **[Entity Validators](#entity-validators)**: Individual entity properties and constraints
- **[Relationship Validators](#relationship-validators)**: Cross-file references and relationships
- **[Business Logic Validators](#business-logic-validators)**: Operational and logical consistency
- **[Accessibility Validators](#accessibility-validators)**: Accessibility feature validation
- **[Fare Validators](#fare-validators)**: Fare system validation
- **[Meta Validators](#meta-validators)**: Feed metadata validation

## Validation Process Flow

1. **File Structure Validation**: Parse files and validate basic CSV structure
2. **Core Validation**: Check required fields, formats, and basic constraints
3. **Entity Validation**: Validate individual records and their properties
4. **Relationship Validation**: Verify cross-file references and foreign keys
5. **Business Logic Validation**: Check operational consistency and logic
6. **Accessibility Validation**: Validate accessibility features
7. **Summary Generation**: Provide overall feed quality assessment

## Notice Severity Levels

- **ERROR** 🔴: Critical violations that break GTFS compliance
- **WARNING** 🟡: Issues that may cause problems but don't break compliance  
- **INFO** 🔵: Informational notices about best practices

---

## Core Validators

Core validators handle fundamental file structure, formatting, and required field validation.

### FileStructureValidator
**Purpose**: Validates the overall structure of GTFS files

**Rules**:
- Validates CSV file format and structure
- Checks for proper file encoding
- Validates file headers and column structure
- Detects malformed CSV files
- Handles empty files appropriately

**Structure Validation**:
- CSV parsing and format validation
- Header row validation
- Data row structure validation
- File encoding detection

**Error Codes**:
- `EmptyFileNotice`
- `CSVParsingFailedNotice`
- `InvalidFileStructureNotice`

### MissingFilesValidator
**Purpose**: Validates presence of required and conditional GTFS files

**Rules**:
- ✅ Required files must exist:
  - `agency.txt` - Transit agency information
  - `stops.txt` - Stop locations  
  - `routes.txt` - Transit route definitions
  - `trips.txt` - Individual trip instances
  - `stop_times.txt` - Stop times for each trip

- 🔄 Conditional requirements:
  - At least one of `calendar.txt` or `calendar_dates.txt` must exist
  - If `translations.txt` exists, `feed_info.txt` is required
  - If `fare_rules.txt` exists, `fare_attributes.txt` is required
  - If `pathways.txt` exists, `levels.txt` is required

**Error Codes**:
- `MissingRequiredFileNotice`
- `MissingCalendarAndCalendarDateFilesNotice`
- `MissingRequiredFileForOptionalFileNotice`

### RequiredFieldValidator
**Purpose**: Validates that all required fields are present and non-empty

**Rules by File**:

#### agency.txt
- **Required**: `agency_name`, `agency_url`, `agency_timezone`
- **Conditional**: `agency_id` (required if multiple agencies exist)

#### stops.txt
- **Required**: `stop_id`
- **Conditional**: `stop_name` (required except for location types 3 and 4)
- **Location-specific**: Coordinates required for certain location types

#### routes.txt
- **Required**: `route_id`, `route_type`
- **Recommended**: `route_short_name` or `route_long_name`

#### trips.txt
- **Required**: `route_id`, `service_id`, `trip_id`

#### stop_times.txt
- **Required**: `trip_id`, `stop_id`, `stop_sequence`
- **Time requirements**: First and last stops need times

#### calendar.txt
- **Required**: All day fields (monday-sunday), `service_id`, `start_date`, `end_date`

**Error Codes**:
- `MissingRequiredFieldNotice`
- `MissingRecommendedFieldNotice`

### FieldFormatValidator
**Purpose**: Validates field formats including URLs, emails, timezones, colors, dates, and times

**Validation Rules**:

#### URL Validation
- Must start with `http://` or `https://`
- Must have valid host component
- Applied to: `agency_url`, `route_url`, `stop_url`

#### Email Validation
- Standard email format validation
- Applied to: `agency_email`

#### Timezone Validation
- Must be valid IANA timezone identifier
- Applied to: `agency_timezone`, `stop_timezone`

#### Color Validation
- 6-digit hexadecimal format (without #)
- Applied to: `route_color`, `route_text_color`

#### Time Validation
- HH:MM:SS format (24-hour, allows >24 for next-day service)
- Applied to: `arrival_time`, `departure_time`

#### Date Validation
- YYYYMMDD format
- Applied to: `start_date`, `end_date`, `date`

**Error Codes**:
- `InvalidURLNotice`
- `InvalidEmailNotice`
- `InvalidTimezoneNotice`
- `InvalidColorFormatNotice`
- `InvalidTimeFormatNotice`
- `InvalidDateFormatNotice`

### Other Core Validators

#### EmptyFileValidator
- Detects completely empty files
- Error: `EmptyFileNotice`

#### DuplicateHeaderValidator
- Finds duplicate column headers within files
- Error: `DuplicateColumnNotice`

#### LeadingTrailingWhitespaceValidator
- Detects leading/trailing whitespace in field values
- Warning: `LeadingOrTrailingWhitespaceNotice`

#### MissingColumnValidator
- Checks for missing required columns
- Error: `MissingColumnNotice`

#### UnknownFileValidator
- Identifies files not defined in GTFS specification
- Info: `UnknownFileNotice`

#### CoordinateValidator
- Validates latitude range: -90.0 to 90.0
- Validates longitude range: -180.0 to 180.0
- Error: `InvalidCoordinateNotice`

#### CurrencyValidator
**Purpose**: Validates currency codes in fare attributes

**Rules**:
- Currency codes must be valid ISO 4217 3-letter codes
- Applied to: `fare_attributes.txt.currency_type`
- Common valid codes: USD, EUR, GBP, CAD, AUD, etc.

**Error Code**: `InvalidCurrencyNotice`

#### DateFormatValidator
**Purpose**: Validates date format consistency

**Rules**:
- Dates must be in YYYYMMDD format
- Applied to: `start_date`, `end_date`, `date` fields
- Validates date range logic (start_date ≤ end_date)

**Error Codes**:
- `InvalidDateFormatNotice`
- `InvalidDateRangeNotice`

#### TimeFormatValidator
**Purpose**: Validates time format consistency

**Rules**:
- Times must be in HH:MM:SS format (24-hour)
- Allows times > 24:00:00 for next-day service
- Applied to: `arrival_time`, `departure_time` fields

**Error Code**: `InvalidTimeFormatNotice`

#### InvalidRowValidator
**Purpose**: Validates CSV row structure and data integrity

**Rules**:
- Checks for malformed CSV rows
- Validates field count matches header count
- Detects encoding issues and parsing errors

**Error Code**: `InvalidRowNotice`

---

## Entity Validators

Entity validators focus on individual records and their specific properties.

### PrimaryKeyValidator
**Purpose**: Validates primary key uniqueness across all files

**Primary Keys by File**:
- `agency.txt`: `agency_id`
- `stops.txt`: `stop_id`
- `routes.txt`: `route_id`
- `trips.txt`: `trip_id`
- `stop_times.txt`: `trip_id` + `stop_sequence`
- `calendar.txt`: `service_id`
- `calendar_dates.txt`: `service_id` + `date`
- `fare_attributes.txt`: `fare_id`
- `fare_rules.txt`: `fare_id` + `route_id` + `origin_id` + `destination_id` + `contains_id`
- `shapes.txt`: `shape_id` + `shape_pt_sequence`
- `frequencies.txt`: `trip_id` + `start_time`
- `transfers.txt`: `from_stop_id` + `to_stop_id`
- `pathways.txt`: `pathway_id`
- `levels.txt`: `level_id`
- `feed_info.txt`: All fields (only one row allowed)

**Error Code**: `DuplicateKeyNotice`

### RouteTypeValidator
**Purpose**: Validates route type values and consistency

**Valid Route Types**:
- **Basic Types (0-12)**:
  - `0` - Tram, Streetcar, Light rail
  - `1` - Subway, Metro
  - `2` - Rail
  - `3` - Bus
  - `4` - Ferry
  - `5` - Cable tram
  - `6` - Aerial lift, Gondola
  - `7` - Funicular
  - `11` - Trolleybus
  - `12` - Monorail

- **Extended Types (100-1799)**: NeTEx/Transmodel extended hierarchy

**Rules**:
- Validates against accepted route type ranges
- Checks consistency between route type and route names
- Warns about uncommon route types
- Identifies suspicious route type combinations

**Error Codes**:
- `InvalidRouteTypeNotice`
- `RouteTypeNameMismatchNotice`
- `UncommonRouteTypeNotice`

### AgencyConsistencyValidator
**Purpose**: Validates agency information consistency

**Rules**:
- Multiple agencies must have unique `agency_id`
- Single agency can omit `agency_id`
- All routes must reference valid agencies
- Consistent agency references across the feed

**Error Codes**:
- `MultipleAgenciesWithoutIdNotice`
- `AgencyReferenceMismatchNotice`

### BikeAllowanceValidator
**Purpose**: Validates `bikes_allowed` field in trips.txt

**Valid Values**:
- `0` - No bike information available
- `1` - Vehicle can accommodate at least one bicycle
- `2` - No bicycles allowed

**Rules**:
- Only accepts values 0, 1, or 2
- Consistency checks with vehicle capacity when available

**Error Code**: `InvalidBikesAllowedNotice`

### StopLocationValidator
**Purpose**: Validates stop location properties

**Rules**:
- Stop coordinates within valid ranges
- Location type consistency with parent stations
- Platform and station relationships
- Entrance and node positioning

**Location Types**:
- `0` - Stop/Platform (default)
- `1` - Station
- `2` - Entrance/Exit
- `3` - Generic Node
- `4` - Boarding Area

**Error Codes**:
- `InvalidStopLocationTypeNotice`
- `StationWithoutPlatformsNotice`
- `PlatformWithoutParentStationNotice`

### StopNameValidator
**Purpose**: Validates stop names for quality, consistency, and accessibility

**Rules**:
- Stop names are required for location types 0, 1, and 2 (stops, stations, entrances)
- Stop names are optional for location types 3 and 4 (nodes, boarding areas)
- Names should be descriptive, meaningful, and user-friendly
- Validates naming conventions and consistency across the feed
- Checks for accessibility and readability issues

**Stop Name Validation Logic**:
- **Required Names**: Stops, stations, and entrances must have names
- **Parent Inheritance**: Child stops can inherit names from parent stations
- **Generic Names**: Warns about generic names like "stop", "station", "platform"
- **Length Validation**: Names should not be excessively long (>50 characters)
- **Character Validation**: Checks for problematic characters and formatting
- **Duplicate Content**: Warns if stop_name and stop_desc are identical
- **Readability**: Flags all-caps names as poor readability
- **Repetition**: Detects repeated words or phrases

**Error Codes**:
- `MissingRequiredStopNameNotice` - Required stop name is missing
- `StopNameMissingButInheritedNotice` - Name missing but can inherit from parent
- `GenericStopNameNotice` - Stop name is too generic
- `StopNameTooLongNotice` - Stop name exceeds reasonable length
- `StopNameWithProblematicCharactersNotice` - Name contains problematic characters
- `StopNameAndDescriptionIdenticalNotice` - Name and description are the same
- `StopNameAllCapsNotice` - Name is in all caps (poor readability)
- `StopNameRepeatedWordsNotice` - Name contains repeated words

### StopTimeHeadsignValidator
**Purpose**: Validates stop time headsigns for consistency and user experience

**Rules**:
- Headsigns should be consistent within trips and across stops
- Headsigns should match route direction and destination
- Validates consistency between trip_headsign and stop_time headsigns
- Checks for logical headsign progression along trip route

**Headsign Validation Logic**:
- **Trip Consistency**: All stops in a trip should have consistent headsigns
- **Direction Logic**: Headsigns should reflect the direction of travel
- **Trip vs Stop Headsigns**: Validates consistency between trip and stop headsigns
- **Progression Logic**: Headsigns should make sense as the trip progresses
- **User Experience**: Headsigns should be clear and helpful for passengers

**Error Codes**:
- `InconsistentStopTimeHeadsignNotice` - Headsigns inconsistent within trip
- `StopTimeHeadsignMismatchNotice` - Stop headsign doesn't match trip headsign
- `InvalidStopTimeHeadsignNotice` - Headsign is invalid or problematic

### RouteNameValidator
**Purpose**: Validates route names for quality, consistency, and best practices

**Rules**:
- At least one of `route_short_name` or `route_long_name` must be provided
- Names should be descriptive, meaningful, and user-friendly
- Validates naming conventions specific to route types
- Checks for consistency and best practices across the feed

**Route Name Validation Logic**:
- **Required Names**: At least one route name field must be non-empty
- **Name Consistency**: Short and long names should not be identical
- **Route Type Conventions**: Validates naming patterns for different route types
- **Naming Best Practices**: Ensures names follow GTFS best practices
- **User Experience**: Names should be clear and helpful for passengers

**Error Codes**:
- `MissingRouteNameNotice` - Both short and long names are empty
- `SameNameAndDescriptionNotice` - Short and long names are identical
- `RouteTypeNameMismatchNotice` - Name doesn't match route type conventions

### DuplicateRouteNameValidator
**Purpose**: Validates route name uniqueness

**Rules**:
- Route names should be unique within the feed
- Handles both short and long names
- Case-insensitive comparison
- Warns about potential confusion for passengers

**Error Codes**:
- `DuplicateRouteNameNotice`

### RouteColorContrastValidator
**Purpose**: Validates color contrast between route colors and text colors

**Rules**:
- Validates WCAG 2.1 AA contrast ratios (4.5:1 for normal text)
- Ensures accessibility compliance for colorblind users
- Checks both `route_color` and `route_text_color` combinations
- Warns about low contrast combinations

**Color Validation**:
- Validates 6-digit hexadecimal format
- Converts to RGB for contrast calculations
- Handles default colors (white/black)

**Error Codes**:
- `InsufficientColorContrastNotice`
- `InvalidRouteColorNotice`

### TripPatternValidator
**Purpose**: Validates trip patterns and stop sequences for consistency and quality

**Rules**:
- Similar trips should follow consistent stop sequences
- Validates stop sequence patterns across trips on the same route
- Checks for logical trip progression and routing
- Identifies potential data quality issues and inconsistencies

**Trip Pattern Validation Logic**:
- **Stop Sequence Analysis**: Groups trips by their stop sequence patterns
- **Pattern Consistency**: Similar trips should have consistent stop sequences
- **Route Pattern Validation**: Validates that trips follow logical route patterns
- **Data Quality Checks**: Identifies trips with unusual or inconsistent patterns
- **Pattern Classification**: Categorizes trips by their stop sequence patterns

**Error Codes**:
- `InconsistentTripPatternNotice` - Trip has inconsistent pattern compared to similar trips
- `InvalidTripPatternNotice` - Trip pattern is invalid or illogical
- `UnusualTripPatternNotice` - Trip pattern is unusual compared to route norm

### TripBlockIdValidator
**Purpose**: Validates trip block ID assignments and consistency for vehicle scheduling

**Rules**:
- Block IDs group trips that are operated by the same vehicle
- All trips within a block should have the same service_id
- Blocks should contain multiple trips for efficient vehicle utilization
- Validates temporal consistency of trips within blocks
- Checks for logical block assignments across routes

**Block Validation Logic**:
- **Service Consistency**: All trips in a block must have the same service_id
- **Route Patterns**: Blocks can span multiple routes (common for interlining)
- **Trip Count**: Blocks should typically contain 2+ trips (single trips flagged as info)
- **Temporal Logic**: Trips within blocks should have logical time progression
- **Performance**: Large blocks (>20 trips) may indicate scheduling issues

**Error Codes**:
- `SingleTripBlockNotice` - Block contains only one trip
- `BlockServiceMismatchNotice` - Trips in block have different service_ids
- `BlockMultipleRoutesNotice` - Block spans multiple routes (info)
- `BlockTooManyTripsNotice` - Block contains excessive number of trips

### ZoneValidator
**Purpose**: Validates fare zone definitions and usage consistency

**Rules**:
- Zone IDs must be consistent between stops.txt and fare_rules.txt
- All zones referenced in fare rules must exist in stops.txt
- Zone assignments should be logical and complete
- Validates zone coverage across the transit network

**Zone Validation Logic**:
- **Zone Definition**: Zones are defined in stops.txt via zone_id field
- **Zone Usage**: Zones are referenced in fare_rules.txt for fare calculations
- **Cross-Reference**: All zones in fare rules must exist in stop definitions
- **Coverage**: Validates that zones provide adequate fare coverage
- **Consistency**: Zone IDs should follow consistent naming conventions

**Error Codes**:
- `OrphanedZoneNotice` - Zone referenced in fare rules but not defined in stops
- `UnusedZoneNotice` - Zone defined in stops but not used in fare rules
- `InvalidZoneIdNotice` - Zone ID format or consistency issues

### CalendarValidator
**Purpose**: Validates that at least one calendar file exists and has data

**Rules**:
- At least one of calendar.txt or calendar_dates.txt must exist
- Calendar files must contain actual data (not just headers)
- Validates basic calendar file structure and content
- Ensures service definitions are available for the feed

**Calendar Validation Logic**:
- **File Presence**: Checks for existence of calendar.txt or calendar_dates.txt
- **Data Content**: Validates that files contain at least one data row
- **Service Definition**: Ensures services can be defined for trips
- **File Structure**: Basic validation of calendar file format

**Error Codes**:
- `MissingCalendarAndCalendarDateFilesNotice` - Neither calendar file exists or has data

### AttributionWithoutRoleValidator
**Purpose**: Validates that attributions have at least one role assigned

**Rules**:
- Each attribution must have at least one role flag set to 1
- Valid roles: `is_producer`, `is_operator`, `is_authority`
- Organization name must be provided
- Attribution ID should be unique if provided

**Role Validation**:
- At least one of the three role fields must be 1
- All role fields should be 0 or 1
- No attribution should have all roles set to 0

**Error Codes**:
- `AttributionWithoutRoleNotice`
- `InvalidAttributionRoleNotice`

---

## Relationship Validators

Relationship validators ensure data consistency across different GTFS files.

### ForeignKeyValidator
**Purpose**: Validates foreign key references between files

**Key Relationships Validated**:

#### Agency References
- `routes.txt.agency_id` → `agency.txt.agency_id`

#### Route References
- `trips.txt.route_id` → `routes.txt.route_id`
- `fare_rules.txt.route_id` → `routes.txt.route_id`

#### Service References
- `trips.txt.service_id` → `calendar.txt.service_id` OR `calendar_dates.txt.service_id`

#### Trip References
- `stop_times.txt.trip_id` → `trips.txt.trip_id`
- `frequencies.txt.trip_id` → `trips.txt.trip_id`

#### Stop References
- `stop_times.txt.stop_id` → `stops.txt.stop_id`
- `stops.txt.parent_station` → `stops.txt.stop_id`
- `transfers.txt.from_stop_id` → `stops.txt.stop_id`
- `transfers.txt.to_stop_id` → `stops.txt.stop_id`
- `pathways.txt.from_stop_id` → `stops.txt.stop_id`
- `pathways.txt.to_stop_id` → `stops.txt.stop_id`

#### Shape References
- `trips.txt.shape_id` → `shapes.txt.shape_id`

#### Fare References
- `fare_rules.txt.fare_id` → `fare_attributes.txt.fare_id`

#### Zone References
- `stops.txt.zone_id` → Self-consistent within feed
- `fare_rules.txt.origin_id` → `stops.txt.zone_id`
- `fare_rules.txt.destination_id` → `stops.txt.zone_id`

**Error Code**: `ForeignKeyViolationNotice`

### StopTimeSequenceValidator
**Purpose**: Validates stop time sequences and timing logic

**Rules**:
- Stop sequences must be unique per trip
- Stop sequences should be increasing (non-decreasing allowed with warning)
- Arrival times should not decrease along trip
- Departure times should not decrease along trip
- Departure time ≥ arrival time at each stop

**Error Codes**:
- `DuplicateStopSequenceNotice`
- `DecreasingStopSequenceNotice`
- `StopTimeDecreasingTimeNotice`
- `ArrivalAfterDepartureNotice`

### ShapeDistanceValidator
**Purpose**: Validates shape distance consistency

**Rules**:
- Shape distances should be increasing along shape
- Distance should be non-negative
- Distance consistency with geographic calculations
- Missing distance warnings when previous distances exist

**Error Codes**:
- `DecreasingOrEqualShapeDistanceNotice`
- `NegativeShapeDistanceNotice`

### ShapeIncreasingDistanceValidator
**Purpose**: Validates that shape distances strictly increase along the shape

**Rules**:
- Each shape point distance must be greater than the previous
- Handles missing distance values appropriately
- Validates distance progression logic
- Ensures proper shape geometry

**Error Codes**:
- `NonIncreasingShapeDistanceNotice`
- `InvalidShapeDistanceNotice`

### StopTimeConsistencyValidator
**Purpose**: Validates stop time consistency across trips

**Rules**:
- Stop times should be consistent for similar trips
- Validates time patterns and intervals
- Checks for logical time progression
- Identifies potential scheduling issues

**Error Codes**:
- `InconsistentStopTimeNotice`
- `InvalidStopTimePatternNotice`

### StopTimeSequenceTimeValidator
**Purpose**: Validates that arrival/departure times are logical within trips

**Rules**:
- Arrival times should not decrease along trip sequence
- Departure times should not decrease along trip sequence
- Departure time ≥ arrival time at each stop
- Handles overnight service (>24:00:00 times)

**Time Validation Logic**:
- Converts times to seconds since midnight for comparison
- Allows for overnight service (times > 86400 seconds)
- Validates logical progression within each trip

**Error Codes**:
- `DecreasingArrivalTimeNotice`
- `DecreasingDepartureTimeNotice`
- `ArrivalAfterDepartureNotice`

### AttributionValidator
**Purpose**: Validates attribution relationships and references

**Rules**:
- Attribution references should be valid
- Attribution IDs should be consistent
- Organization information should be complete
- Attribution relationships should be logical

**Error Codes**:
- `InvalidAttributionReferenceNotice`
- `InconsistentAttributionNotice`

---

## Business Logic Validators

Business logic validators check operational feasibility and real-world consistency.

### TravelSpeedValidator
**Purpose**: Validates travel speeds between stops are reasonable

**Speed Limits by Route Type**:
- **Bus/Trolleybus (3, 11)**: 150 km/h
- **Rail/Subway (1, 2)**: 500 km/h  
- **Tram/Light Rail (0)**: 100 km/h
- **Ferry (4)**: 100 km/h
- **Cable systems (5, 6, 7)**: 50 km/h
- **Monorail (12)**: 300 km/h

**Rules**:
- Calculate speed from geographic distance and travel time
- Flag speeds exceeding reasonable limits for transport mode
- Consider both inter-stop and intra-trip speed patterns

**Error Codes**:
- `ExcessiveTravelSpeedNotice`
- `ImpossibleTravelTimeNotice`

### ServiceConsistencyValidator
**Purpose**: Validates service definitions and usage

**Rules**:
- Services must be active on at least one day
- Date ranges must be valid (start_date ≤ end_date)
- Services should not be too old (> 2 years) or too future (> 2 years)
- All defined services should be used by at least one trip
- Service exceptions in calendar_dates should reference existing services

**Error Codes**:
- `ServiceNeverActiveNotice`
- `InvalidDateRangeNotice`
- `VeryOldServiceNotice`
- `VeryFutureServiceNotice`
- `UnusedServiceNotice`
- `ServiceWithoutTripsNotice`

### FeedExpirationDateValidator
**Purpose**: Validates feed expiration and freshness

**Rules**:
- Feed should not be expired based on feed_info.txt dates
- Service should be available for reasonable future periods
- Validates both feed-level and service-level expiration
- Checks calendar-based expiration when feed_info.txt is not available

**Expiration Validation**:
- Feed end date should be in the future
- Service coverage should extend at least 7 days ahead
- Warns about feeds expiring soon
- Validates service date ranges

**Error Codes**:
- `ExpiredFeedNotice`
- `FeedExpiringSoonNotice`
- `InsufficientServiceCoverageNotice`

### DateTripsValidator
**Purpose**: Validates that trips exist for the next 7 days with majority service coverage

**Rules**:
- At least 50% of services should be active in the next 7 days
- Service should be available for the next 30 days (warning level)
- Validates both calendar.txt and calendar_dates.txt service definitions
- Checks for adequate trip coverage on active service days

**Service Coverage Validation**:
- Analyzes service patterns for next 7 and 30 days
- Considers calendar exceptions (additions/removals)
- Validates trip availability on active service days
- Ensures reasonable service continuity

**Error Codes**:
- `NoServiceDefinedNotice`
- `InsufficientServiceCoverageNotice`
- `NoTripsForNext7DaysNotice`
- `NoTripsForNext30DaysNotice`

### GeospatialValidator
**Purpose**: Validates geographic data quality, consistency, and spatial relationships

**Rules**:
- Validates coordinate accuracy and precision across all geographic data
- Ensures spatial relationships between stops are logical and consistent
- Validates shape geometry and distance calculations
- Analyzes feed coverage and stop clustering patterns
- Checks for geographic data quality issues and potential errors

**Geographic Validation Logic**:

#### Coordinate Validation
- **Latitude Range**: Must be between -90.0 and 90.0 degrees
- **Longitude Range**: Must be between -180.0 and 180.0 degrees
- **Suspicious Coordinates**: Flags coordinates at exactly (0, 0) as potential errors
- **Coordinate Precision**: Validates decimal precision for accuracy (minimum 4 decimal places for ~11m precision)

#### Feed Coverage Analysis
- **Bounding Box Calculation**: Determines geographic extent of the feed
- **Large Coverage Detection**: Warns if feed spans >1000km in any direction (potential data errors)
- **Small Coverage Detection**: Warns if feed spans <1km in any direction (precision issues)
- **Coverage Consistency**: Validates that all stops and shapes fall within reasonable bounds

#### Stop Spatial Relationships
- **Parent-Child Distance**: Validates child stops are within 500m of parent stations
- **Duplicate Detection**: Identifies stops within 10m of each other (potential duplicates)
- **Stop Clustering**: Analyzes stop density patterns and clustering behavior
- **High Density Areas**: Flags areas with >20 stops in 500m radius

#### Shape Geometry Validation
- **Shape Bounds**: Ensures shape points fall within feed geographic bounds
- **Segment Length**: Validates shape segments are not unreasonably long (>50km)
- **Distance Consistency**: Cross-checks shape_dist_traveled against geographic calculations
- **Shape Quality**: Validates shape point density and distribution

#### Distance Calculations
- **Haversine Formula**: Uses great-circle distance calculations for accuracy
- **Geographic Distance**: Calculates actual geographic distances between points
- **Distance Tolerance**: Allows 20% tolerance for shape distance vs. geographic distance
- **Precision Validation**: Ensures coordinate precision is sufficient for accurate calculations

**Error Codes**:
- `InvalidLatitudeNotice` - Latitude outside valid range (-90 to 90)
- `InvalidLongitudeNotice` - Longitude outside valid range (-180 to 180)
- `SuspiciousCoordinateNotice` - Coordinates at (0, 0) or other suspicious values
- `VeryLargeFeedCoverageNotice` - Feed coverage exceeds 1000km in any direction
- `VerySmallFeedCoverageNotice` - Feed coverage less than 1km in any direction
- `ChildStationTooFarFromParentNotice` - Child stop >500m from parent station
- `VeryCloseStopsNotice` - Stops within 10m of each other (potential duplicates)
- `ShapePointOutsideFeedBoundsNotice` - Shape point outside feed geographic bounds
- `UnreasonablyLongShapeSegmentNotice` - Shape segment >50km (missing points)
- `ShapeDistanceInconsistentWithGeographyNotice` - Shape distance doesn't match geographic distance
- `HighStopDensityAreaNotice` - Area with >20 stops in 500m radius
- `LowStopClusteringNotice` - Feed has very few stop clusters
- `InsufficientCoordinatePrecisionNotice` - Coordinate precision <4 decimal places
- `GeospatialSummaryNotice` - Summary of geospatial analysis results

### FrequencyValidator
**Purpose**: Validates frequency-based service definitions and scheduling logic

**Rules**:
- `start_time` < `end_time` for each frequency period
- `headway_secs` must be greater than 0
- Reasonable headway ranges for operational feasibility
- No overlapping frequency periods for the same trip
- Trip references must exist in trips.txt
- Validates exact_times field values

**Frequency Validation Logic**:

#### Time Range Validation
- **Start/End Time Logic**: Start time must be before end time for each frequency period
- **Time Format**: Validates HH:MM:SS format and converts to seconds since midnight
- **Overnight Service**: Handles times > 24:00:00 for next-day service

#### Headway Validation
- **Minimum Headway**: Headways must be > 0 seconds
- **Reasonable Range**: Headways should be between 30 seconds and 4 hours
- **Operational Feasibility**: Very short headways (<30s) or very long headways (>4h) flagged
- **Service Quality**: Validates headways are practical for passenger service

#### Overlap Detection
- **Period Sorting**: Sorts frequency periods by start time for overlap detection
- **Overlap Logic**: Ensures no frequency periods overlap for the same trip
- **Service Continuity**: Validates seamless frequency-based service coverage

#### Exact Times Validation
- **Field Values**: exact_times must be 0 (frequency-based) or 1 (schedule-based)
- **Service Type**: Distinguishes between frequency-based and schedule-based trips
- **Consistency**: Validates field consistency across frequency records

**Error Codes**:
- `InvalidFrequencyTimeRangeNotice` - Start time >= end time
- `InvalidHeadwayNotice` - Headway <= 0 seconds
- `UnreasonableHeadwayNotice` - Headway < 30s or > 4 hours
- `OverlappingFrequencyNotice` - Frequency periods overlap for same trip
- `InvalidExactTimesNotice` - exact_times not 0 or 1
- `ForeignKeyViolationNotice` - Trip reference doesn't exist in trips.txt

### NetworkTopologyValidator
**Purpose**: Analyzes network connectivity and topology

**Rules**:
- Connected route network analysis
- Isolated stops detection
- Transfer connectivity validation
- Route continuity checks
- Network graph analysis for connectivity

**Topology Analysis**:
- Identifies stops without connections
- Validates transfer network completeness
- Checks for disconnected route segments
- Analyzes network graph properties

**Error Codes**:
- `IsolatedStopNotice`
- `DisconnectedRouteNotice`
- `PoorNetworkConnectivityNotice`

---

## Accessibility Validators

Accessibility validators ensure compliance with accessibility standards and requirements.

### PathwayValidator
**Purpose**: Validates pathway definitions for accessibility compliance

**Pathway Modes**:
- `1` - Walkway
- `2` - Stairs  
- `3` - Moving sidewalk/travelator
- `4` - Escalator
- `5` - Elevator
- `6` - Fare gate, payment gate
- `7` - Exit gate

**Rules**:
- Valid pathway mode values (1-7)
- Bidirectional consistency (`is_bidirectional` field)
- Mode-specific requirements:
  - Stairs: `stair_count` should be provided
  - Escalators: direction considerations
  - Elevators: level connectivity
- Numeric field validation (length, traversal_time, slope)
- Gate bidirectionality warnings (gates are typically unidirectional)

**Error Codes**:
- `InvalidPathwayModeNotice`
- `UnexpectedBidirectionalGateNotice`
- `MissingPathwayStairCountNotice`
- `InvalidPathwayLengthNotice`

### LevelValidator
**Purpose**: Validates level definitions for multi-level stations and accessibility compliance

**Rules**:
- Level IDs must be unique across the feed
- Level index values should be reasonable and consistent
- Level names should be descriptive and accessible
- Levels should be referenced by stops or pathways
- Validates level hierarchy and accessibility features

**Level Validation Logic**:

#### Level Index Validation
- **Range Validation**: Level indices should be between -50 and +50 (reasonable bounds)
- **Uniqueness**: Each level index should be unique within the feed
- **Hierarchy Logic**: Lower indices typically represent lower levels (basements, etc.)
- **Accessibility**: Validates level indices support accessibility navigation

#### Level Name Validation
- **Required Field**: level_name is recommended for accessibility compliance
- **Descriptive Names**: Names should be clear and helpful for navigation
- **Accessibility**: Names should support accessibility features and wayfinding
- **Consistency**: Level naming should follow consistent conventions

#### Usage Validation
- **Stop References**: Validates levels are referenced in stops.txt
- **Pathway References**: Validates levels are referenced in pathways.txt
- **Cross-Reference**: Ensures level definitions are actually used
- **Orphaned Levels**: Flags levels that are defined but not referenced

#### Accessibility Compliance
- **Multi-Level Stations**: Validates proper level definitions for complex stations
- **Navigation Support**: Ensures levels support passenger navigation
- **Accessibility Features**: Validates level information supports accessibility needs
- **Wayfinding**: Ensures level data supports passenger wayfinding

**Error Codes**:
- `UnreasonableLevelIndexNotice` - Level index outside reasonable range (-50 to +50)
- `MissingRecommendedFieldNotice` - level_name field is missing (recommended)
- `DuplicateLevelIndexNotice` - Multiple levels have the same index value
- `UnusedLevelNotice` - Level is defined but not referenced by stops or pathways

---

## Fare Validators

Fare validators ensure fare system definitions are complete and consistent.

### FareValidator
**Purpose**: Validates fare system definitions and rules

**Fare Attributes Validation**:
- **Price**: Non-negative, reasonable precision (typically 2 decimal places)
- **Payment Method**: 
  - `0` - Fare is paid on board
  - `1` - Fare must be paid before boarding
- **Transfers**: 
  - `0` - No transfers permitted
  - `1` - Passengers may transfer once
  - `2` - Passengers may transfer twice
  - (empty) - Unlimited transfers permitted
- **Transfer Duration**: Reasonable time limits (typically 30 minutes to 24 hours)

**Fare Rules Validation**:
- At least one of route_id, origin_id, destination_id, or contains_id must be specified
- All referenced IDs must exist in appropriate files
- Fare rule completeness and coverage

**Usage Analysis**:
- Unused fare attributes detection
- Fare rule coverage analysis
- Price consistency checks

**Error Codes**:
- `InvalidFarePriceNotice`
- `InvalidPaymentMethodNotice`
- `InvalidTransfersNotice`
- `InvalidTransferDurationNotice`
- `EmptyFareRuleNotice`
- `UnusedFareAttributeNotice`

---

## Meta Validators

Meta validators handle feed-level metadata and information.

### FeedInfoValidator
**Purpose**: Validates feed metadata and information

**Required Fields**:
- `feed_publisher_name`: Name of the organization that publishes the feed
- `feed_publisher_url`: URL of the feed publisher
- `feed_lang`: Default language code (ISO 639-1)

**Optional Fields**:
- `feed_start_date`: First date of service in the feed
- `feed_end_date`: Last date of service in the feed  
- `feed_version`: Version identifier for the feed
- `feed_contact_email`: Contact email for feed issues
- `feed_contact_url`: Contact URL for feed information

**Rules**:
- Only one feed_info entry is allowed
- URL format validation for publisher and contact URLs
- Language code validation (2-letter ISO 639-1 codes)
- Date range validation (start_date ≤ end_date)
- Email format validation for contact email
- Feed expiration warnings

**Error Codes**:
- `MultipleFeedInfoEntriesNotice`
- `InvalidFeedLanguageNotice`
- `ExpiredFeedNotice`
- `FutureFeedNotice`
- `InvalidFeedContactInfoNotice`

---

## Configuration Options

The validator supports various configuration options:

### Global Configuration
```go
type Config struct {
    CountryCode    string        // Country-specific validation rules (e.g., "US", "GB")
    CurrentDate    time.Time     // Reference date for expiration checks
    MaxMemory      int64         // Memory usage limit in bytes
    ParallelWorkers int          // Number of concurrent validation workers
    ValidationMode ValidationMode // Validation thoroughness level
}
```

### Validation Modes

#### Performance Mode
- Faster validation with essential checks only
- Reduced memory usage
- Suitable for large feeds or CI environments
- **Enabled Validators**:
  - Core validators (file structure, required fields, formats)
  - Relationship validators (foreign keys, sequences)
  - Meta validators (feed info)
  - Limited to 50 notices per type

#### Default Mode
- Balanced validation coverage and performance
- Most common validation use case
- Recommended for general use
- **Enabled Validators**:
  - All core validators
  - All entity validators
  - All relationship validators
  - All business validators (except expensive ones)
  - All accessibility validators
  - All fare validators
  - All meta validators
  - Limited to 100 notices per type

#### Comprehensive Mode
- Most thorough validation
- All available checks enabled
- Suitable for feed certification and detailed analysis
- **Enabled Validators**:
  - All validators including expensive ones:
    - GeospatialValidator (geographic analysis)
    - NetworkTopologyValidator (network connectivity)
    - DateTripsValidator (service coverage analysis)
  - Limited to 1000 notices per type

### Country-Specific Rules
Different countries may have specific GTFS requirements:
- **US**: FTA compliance requirements
- **EU**: European accessibility standards
- **Canada**: Bilingual requirements
- **Australia**: Australian standards

---

## Error Handling and Notice Types

### Notice Structure
Each validation notice contains:
- **Code**: Unique identifier for the validation rule
- **Severity**: ERROR, WARNING, or INFO
- **Message**: Human-readable description
- **Context**: File, row, and field information
- **Suggestion**: Recommended fix when applicable

### User-Friendly Error Descriptions

The validator provides detailed, user-friendly descriptions for each error type to help feed producers understand and fix issues. Here are examples of the enhanced descriptions:

#### Core Validation Errors

**Missing Required File**
- **Description**: "A required GTFS file is missing from the feed. This file is essential for GTFS compliance and must be present."
- **Impact**: Feed will not be accepted by GTFS consumers
- **Fix**: Add the missing file with proper content and structure

**Missing Required Field**
- **Description**: "A required field is missing from a GTFS file. This field is mandatory according to the GTFS specification."
- **Impact**: Data integrity issues, potential feed rejection
- **Fix**: Add the missing field value to the specified row

**Invalid Date Format**
- **Description**: "A date field contains an invalid date format. Dates must be in YYYYMMDD format (e.g., 20231225 for December 25, 2023)."
- **Impact**: Date parsing errors, service scheduling issues
- **Fix**: Ensure all dates follow YYYYMMDD format

**Invalid Time Format**
- **Description**: "A time field contains an invalid time format. Times must be in HH:MM:SS format (e.g., 14:30:00 for 2:30 PM)."
- **Impact**: Time parsing errors, trip scheduling issues
- **Fix**: Ensure all times follow HH:MM:SS format (24-hour clock)

#### Entity Validation Errors

**Duplicate Primary Key**
- **Description**: "A record has a duplicate primary key. Each record must have a unique identifier to maintain data integrity."
- **Impact**: Data conflicts, potential feed rejection
- **Fix**: Ensure each record has a unique primary key value

**Invalid Route Type**
- **Description**: "A route type field contains an invalid route type. Route types must be valid GTFS route type codes."
- **Impact**: Route classification errors, consumer confusion
- **Fix**: Use valid route type codes (0-12 for basic types, 100-1799 for extended types)

#### Business Logic Errors

**Excessive Travel Speed**
- **Description**: "Travel speed between stops is unrealistically fast for the transport mode. This may indicate data errors or missing stops."
- **Impact**: Unrealistic trip planning, passenger confusion
- **Fix**: Check for missing intermediate stops or correct travel times

**Feed Expiration**
- **Description**: "The feed has expired or will expire soon. Feeds should be updated regularly to provide current service information."
- **Impact**: Outdated information, potential service disruptions
- **Fix**: Update feed with current service dates and republish

#### Accessibility Errors

**Poor Color Contrast**
- **Description**: "Route colors have insufficient contrast for accessibility compliance. This affects colorblind users and accessibility standards."
- **Impact**: Accessibility compliance issues, poor user experience
- **Fix**: Choose route colors with sufficient contrast ratios (WCAG 2.1 AA compliant)

**Missing Pathway Information**
- **Description**: "Pathway information is missing or incomplete. This affects accessibility navigation in multi-level stations."
- **Impact**: Accessibility compliance issues, navigation difficulties
- **Fix**: Add complete pathway definitions for station navigation

### Comprehensive Error Description Reference

#### File Structure Errors

**Empty File**
- **Description**: "A GTFS file is completely empty (no data rows). Empty files may indicate data export issues or missing content."
- **Impact**: File parsing errors, incomplete feed information
- **Fix**: Ensure the file contains proper data or remove if not needed

**Unknown File**
- **Description**: "An unknown file is present in the GTFS feed. Only standard GTFS files should be included."
- **Impact**: Potential confusion, non-standard feed structure
- **Fix**: Remove non-GTFS files or rename to follow GTFS conventions

**Duplicate Column Headers**
- **Description**: "Duplicate column headers found in a GTFS file. Each column should have a unique header name."
- **Impact**: Data parsing errors, column mapping issues
- **Fix**: Remove duplicate columns or rename to make them unique

#### Data Quality Errors

**Leading/Trailing Whitespace**
- **Description**: "Field values contain leading or trailing whitespace. This can cause data parsing issues and inconsistencies."
- **Impact**: Data matching problems, potential validation failures
- **Fix**: Trim whitespace from all field values

**Invalid URL Format**
- **Description**: "A URL field contains an invalid URL format. URLs must be properly formatted with http:// or https:// protocol."
- **Impact**: Broken links, poor user experience
- **Fix**: Ensure URLs include proper protocol and are valid

**Invalid Email Format**
- **Description**: "An email field contains an invalid email address format. Email addresses must follow standard email format."
- **Impact**: Contact information issues, communication problems
- **Fix**: Use valid email address format (e.g., contact@agency.com)

#### Geographic Data Errors

**Invalid Coordinates**
- **Description**: "Coordinates are outside valid ranges. Latitude must be between -90 and 90, longitude between -180 and 180."
- **Impact**: Mapping errors, location display issues
- **Fix**: Correct coordinate values to valid ranges

**Suspicious Coordinates**
- **Description**: "Coordinates appear to be placeholder or error values (e.g., 0,0). This may indicate data import issues."
- **Impact**: Incorrect location display, navigation problems
- **Fix**: Replace with actual geographic coordinates

**Very Close Stops**
- **Description**: "Stops are located very close together (within 10 meters). This may indicate duplicate stops or data errors."
- **Impact**: Confusion for passengers, redundant data
- **Fix**: Review and consolidate duplicate stops or correct coordinates

#### Service and Schedule Errors

**Service Never Active**
- **Description**: "A service is defined but never active on any day. This creates unused service definitions."
- **Impact**: Data bloat, potential confusion
- **Fix**: Remove unused services or add active days

**Invalid Date Range**
- **Description**: "Service start date is after end date. Date ranges must be logical with start date before end date."
- **Impact**: Service scheduling errors, date parsing issues
- **Fix**: Correct the date range order

**Overlapping Frequencies**
- **Description**: "Frequency periods overlap for the same trip. Each time period should be distinct and non-overlapping."
- **Impact**: Scheduling conflicts, service planning issues
- **Fix**: Adjust frequency periods to eliminate overlaps

#### Fare System Errors

**Invalid Fare Price**
- **Description**: "Fare price is negative or has excessive precision. Prices should be non-negative with reasonable decimal places."
- **Impact**: Fare calculation errors, payment issues
- **Fix**: Use non-negative prices with 2 decimal places maximum

**Empty Fare Rule**
- **Description**: "A fare rule has no qualifying conditions specified. At least one condition must be defined."
- **Impact**: Ambiguous fare application, pricing confusion
- **Fix**: Add qualifying conditions (route, origin, destination, or zone)

**Unused Fare Attribute**
- **Description**: "A fare attribute is defined but not used in any fare rules. This creates orphaned fare definitions."
- **Impact**: Data bloat, potential confusion
- **Fix**: Remove unused fare attributes or create corresponding fare rules

#### Accessibility and Compliance Errors

**Insufficient Color Contrast**
- **Description**: "Route colors have insufficient contrast for accessibility compliance. This affects colorblind users."
- **Impact**: Accessibility compliance issues, poor user experience
- **Fix**: Choose colors with WCAG 2.1 AA contrast ratios (4.5:1 minimum)

**Missing Level Information**
- **Description**: "Level information is missing for multi-level stations. This affects accessibility navigation."
- **Impact**: Accessibility compliance issues, navigation difficulties
- **Fix**: Add complete level definitions for multi-level stations

**Invalid Pathway Mode**
- **Description**: "Pathway mode is invalid or not specified. Pathway modes must be valid GTFS pathway type codes."
- **Impact**: Accessibility navigation issues, compliance problems
- **Fix**: Use valid pathway mode codes (1-7)

#### Network and Connectivity Errors

**Isolated Stop**
- **Description**: "A stop cannot be reached by any trip. This creates disconnected network elements."
- **Impact**: Network connectivity issues, passenger confusion
- **Fix**: Add trips serving the stop or remove if not needed

**Disconnected Route**
- **Description**: "A route has no connections to other routes. This may indicate network planning issues."
- **Impact**: Limited passenger options, poor network design
- **Fix**: Review network design and add transfer connections

**Poor Network Connectivity**
- **Description**: "The network has poor overall connectivity. This affects passenger journey planning and accessibility."
- **Impact**: Limited travel options, poor user experience
- **Fix**: Improve network connectivity through better transfer design

### Comparison with Other GTFS Validators

The GTFS validator provides comprehensive error descriptions that compare favorably with other validation tools:

#### vs. Google's GTFS Validator
- **Similar Features**: Both provide clear error messages and severity levels
- **Enhanced Descriptions**: Our validator includes detailed impact analysis and specific fix suggestions
- **Comprehensive Coverage**: Covers more validation scenarios with detailed explanations

#### vs. Java GTFS Validator
- **User-Friendly Language**: More accessible language for non-technical users
- **Actionable Fixes**: Specific, actionable recommendations for each error type
- **Impact Assessment**: Clear explanation of how each error affects feed quality and usability

#### vs. Other Open Source Validators
- **Detailed Context**: Provides file, row, and field-specific context for each error
- **Best Practices**: Includes recommendations based on GTFS best practices
- **Accessibility Focus**: Enhanced descriptions for accessibility compliance issues

#### Key Advantages
- **Comprehensive Error Coverage**: 55+ validators with detailed error descriptions
- **User-Friendly Language**: Clear, non-technical explanations
- **Actionable Recommendations**: Specific steps to fix each issue
- **Impact Assessment**: Understanding of how errors affect feed quality
- **Best Practice Guidance**: Recommendations based on industry standards

### Common Notice Patterns

#### File-Level Notices
- Apply to entire files
- Examples: EmptyFileNotice, MissingRequiredFileNotice

#### Field-Level Notices  
- Apply to specific fields within records
- Examples: MissingRequiredFieldNotice, InvalidFieldFormatNotice

#### Cross-File Notices
- Apply to relationships between files
- Examples: ForeignKeyViolationNotice, UnusedServiceNotice

#### Business Logic Notices
- Apply to operational feasibility
- Examples: ExcessiveTravelSpeedNotice, ImpossibleTravelTimeNotice

---

## Best Practices for Feed Producers

### Essential Validations
1. **Always validate** basic file structure and required fields
2. **Check foreign key relationships** between files
3. **Verify coordinate accuracy** and reasonable geographic coverage
4. **Validate service date ranges** are current and appropriate
5. **Test route connectivity** and transfer relationships

### Performance Considerations
1. **Large feeds** should use Performance mode for faster validation
2. **Memory-constrained environments** should set appropriate MaxMemory limits
3. **CI/CD pipelines** benefit from parallel processing configuration

### Quality Assurance
1. **Regular validation** during feed development
2. **Comprehensive validation** before publishing
3. **Monitor validation reports** for trends and patterns
4. **Address ERROR-level issues** before feed publication
5. **Consider WARNING-level issues** for feed quality improvement

---

## Validation Report Format

The validator generates structured reports with:

### Summary Information
- Feed statistics (routes, trips, stops, etc.)
- Validation time and performance metrics  
- Overall validation status (PASSED/FAILED)

### Notice Details
- Grouped by validation rule and severity
- Sample instances with file/row context
- Total count per notice type

### Recommendations
- Prioritized list of issues to address
- Suggested fixes and improvements
- Links to relevant GTFS documentation

This comprehensive validation ensures GTFS feeds meet specification requirements and operational feasibility standards for public transit data.

---

## Complete Validator Coverage Summary

### Core Validators (12 validators)
1. **FileStructureValidator** - CSV structure and format validation
2. **MissingFilesValidator** - Required file presence validation
3. **EmptyFileValidator** - Empty file detection
4. **UnknownFileValidator** - Unknown file identification
5. **DuplicateHeaderValidator** - Duplicate column header detection
6. **MissingColumnValidator** - Missing required column validation
7. **RequiredFieldValidator** - Required field presence validation
8. **FieldFormatValidator** - Field format validation (URL, email, etc.)
9. **TimeFormatValidator** - Time format validation
10. **DateFormatValidator** - Date format validation
11. **CoordinateValidator** - Coordinate range validation
12. **CurrencyValidator** - Currency code validation
13. **DuplicateKeyValidator** - Primary key uniqueness validation
14. **InvalidRowValidator** - CSV row structure validation
15. **LeadingTrailingWhitespaceValidator** - Whitespace validation

### Entity Validators (19 validators)
1. **PrimaryKeyValidator** - Primary key uniqueness validation
2. **CalendarValidator** - Calendar definition validation
3. **AgencyConsistencyValidator** - Agency consistency validation
4. **RouteConsistencyValidator** - Route consistency validation
5. **ServiceValidationValidator** - Service validation
6. **StopLocationValidator** - Stop location hierarchy validation
7. **CalendarConsistencyValidator** - Calendar consistency validation
8. **ShapeValidator** - Shape definition validation
9. **ZoneValidator** - Zone definition validation
10. **RouteNameValidator** - Route name validation
11. **TripPatternValidator** - Trip pattern validation
12. **DuplicateRouteNameValidator** - Route name uniqueness validation
13. **RouteColorContrastValidator** - Color contrast validation
14. **StopNameValidator** - Stop name validation
15. **BikesAllowanceValidator** - Bike allowance validation
16. **AttributionWithoutRoleValidator** - Attribution role validation
17. **TripBlockIdValidator** - Trip block ID validation
18. **StopTimeHeadsignValidator** - Stop time headsign validation
19. **RouteTypeValidator** - Route type validation

### Relationship Validators (8 validators)
1. **ForeignKeyValidator** - Foreign key reference validation
2. **StopTimeSequenceValidator** - Stop time sequence validation
3. **StopTimeSequenceTimeValidator** - Stop time logical ordering validation
4. **ShapeDistanceValidator** - Shape distance validation
5. **StopTimeConsistencyValidator** - Stop time consistency validation
6. **AttributionValidator** - Attribution relationship validation
7. **RouteConsistencyValidator** - Route relationship validation
8. **ShapeIncreasingDistanceValidator** - Shape distance progression validation

### Business Logic Validators (12 validators)
1. **TripUsabilityValidator** - Trip usability validation
2. **TravelSpeedValidator** - Travel speed validation
3. **BlockOverlappingValidator** - Block overlap validation
4. **FrequencyValidator** - Frequency definition validation
5. **OverlappingFrequencyValidator** - Frequency overlap validation
6. **TransferValidator** - Transfer definition validation
7. **TransferTimingValidator** - Transfer timing validation
8. **FeedExpirationDateValidator** - Feed expiration validation
9. **ServiceConsistencyValidator** - Service consistency validation
10. **ServiceCalendarValidator** - Service calendar validation
11. **ScheduleConsistencyValidator** - Schedule consistency validation
12. **GeospatialValidator** - Geographic data validation (expensive)
13. **NetworkTopologyValidator** - Network connectivity validation (expensive)
14. **DateTripsValidator** - Service coverage validation (expensive)

### Accessibility Validators (2 validators)
1. **PathwayValidator** - Pathway definition validation
2. **LevelValidator** - Level definition validation

### Fare Validators (1 validator)
1. **FareValidator** - Fare system validation

### Meta Validators (1 validator)
1. **FeedInfoValidator** - Feed metadata validation