package notice

import (
	"fmt"
	"strings"
	"sync"
)

// Notice is the interface for all validation notices
type Notice interface {
	Code() string
	Severity() SeverityLevel
	Context() map[string]interface{}
}

// BaseNotice provides common functionality for all notices
type BaseNotice struct {
	code     string
	severity SeverityLevel
	context  map[string]interface{}
}

// NewBaseNotice creates a new base notice
func NewBaseNotice(code string, severity SeverityLevel, context map[string]interface{}) *BaseNotice {
	return &BaseNotice{
		code:     code,
		severity: severity,
		context:  context,
	}
}

// Code returns the notice code
func (n *BaseNotice) Code() string {
	return n.code
}

// Severity returns the notice severity
func (n *BaseNotice) Severity() SeverityLevel {
	return n.severity
}

// Context returns the notice context
func (n *BaseNotice) Context() map[string]interface{} {
	return n.context
}

// GetCode generates a code from a notice type name
func GetCode(typeName string) string {
	// Convert from CamelCase to snake_case
	// Remove "Notice" suffix if present
	name := strings.TrimSuffix(typeName, "Notice")
	
	var result []rune
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	
	return strings.ToLower(string(result))
}

// NoticeContainer holds all notices generated during validation
type NoticeContainer struct {
	notices        []Notice
	noticeCounts   map[string]int
	maxPerType     int
	mutex          sync.RWMutex
}

// NewNoticeContainer creates a new notice container
func NewNoticeContainer() *NoticeContainer {
	return &NoticeContainer{
		notices:      make([]Notice, 0),
		noticeCounts: make(map[string]int),
		maxPerType:   100, // Default limit
	}
}

// NewNoticeContainerWithLimit creates a new notice container with custom limits
func NewNoticeContainerWithLimit(maxPerType int) *NoticeContainer {
	return &NoticeContainer{
		notices:      make([]Notice, 0),
		noticeCounts: make(map[string]int),
		maxPerType:   maxPerType,
	}
}

// AddNotice adds a notice to the container with optional limiting
func (nc *NoticeContainer) AddNotice(notice Notice) {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	
	code := notice.Code()
	
	// Check if we've hit the limit for this notice type
	if nc.maxPerType > 0 && nc.noticeCounts[code] >= nc.maxPerType {
		return // Skip adding more notices of this type
	}
	
	nc.notices = append(nc.notices, notice)
	nc.noticeCounts[code]++
}

// SetMaxNoticesPerType sets the maximum number of notices per type
func (nc *NoticeContainer) SetMaxNoticesPerType(max int) {
	nc.maxPerType = max
}

// GetNotices returns all notices
func (nc *NoticeContainer) GetNotices() []Notice {
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	// Return a copy to avoid data races
	result := make([]Notice, len(nc.notices))
	copy(result, nc.notices)
	return result
}

// GetNoticesByCode returns notices filtered by code
func (nc *NoticeContainer) GetNoticesByCode(code string) []Notice {
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	var filtered []Notice
	for _, n := range nc.notices {
		if n.Code() == code {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

// GetNoticesBySeverity returns notices filtered by severity
func (nc *NoticeContainer) GetNoticesBySeverity(severity SeverityLevel) []Notice {
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	var filtered []Notice
	for _, n := range nc.notices {
		if n.Severity() == severity {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

// CountBySeverity returns the count of notices by severity level
func (nc *NoticeContainer) CountBySeverity() map[SeverityLevel]int {
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	counts := make(map[SeverityLevel]int)
	for _, n := range nc.notices {
		counts[n.Severity()]++
	}
	return counts
}

// HasErrors returns true if there are any ERROR level notices
func (nc *NoticeContainer) HasErrors() bool {
	nc.mutex.RLock()
	defer nc.mutex.RUnlock()
	for _, n := range nc.notices {
		if n.Severity() == ERROR {
			return true
		}
	}
	return false
}

// String returns a string representation of the container
func (nc *NoticeContainer) String() string {
	counts := nc.CountBySeverity()
	return fmt.Sprintf("NoticeContainer{errors: %d, warnings: %d, infos: %d}",
		counts[ERROR], counts[WARNING], counts[INFO])
}