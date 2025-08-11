package notice

// SeverityLevel represents the severity of a validation notice
type SeverityLevel int

const (
	// INFO - for items that do not affect the feed's quality
	INFO SeverityLevel = iota
	// WARNING - for items that affect quality but aren't required by spec
	WARNING
	// ERROR - for items that violate GTFS spec requirements
	ERROR
)

// String returns the string representation of the severity level
func (s SeverityLevel) String() string {
	switch s {
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
