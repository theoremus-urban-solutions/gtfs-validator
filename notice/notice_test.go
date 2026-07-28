package notice

import (
	"testing"
)

func TestSeverityLevel(t *testing.T) {
	tests := []struct {
		severity SeverityLevel
		string   string
	}{
		{ERROR, "ERROR"},
		{WARNING, "WARNING"},
		{INFO, "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.string, func(t *testing.T) {
			if tt.severity.String() != tt.string {
				t.Errorf("Expected %s, got %s", tt.string, tt.severity.String())
			}
		})
	}
}

func TestNoticeContainer(t *testing.T) {
	container := NewNoticeContainer()

	// Test empty container
	if len(container.GetNotices()) != 0 {
		t.Errorf("Expected 0 notices, got %d", len(container.GetNotices()))
	}

	// Test adding notices
	notice1 := NewBaseNotice("test_error", ERROR, map[string]interface{}{
		"filename":  "agency.txt",
		"rowNumber": 2,
	})

	container.AddNotice(notice1)

	if len(container.GetNotices()) != 1 {
		t.Errorf("Expected 1 notice, got %d", len(container.GetNotices()))
	}

	// Test getting notices
	notices := container.GetNotices()
	if len(notices) != 1 {
		t.Errorf("Expected 1 notice in list, got %d", len(notices))
	}

	if notices[0].Code() != "test_error" {
		t.Errorf("Expected notice code 'test_error', got %s", notices[0].Code())
	}
	if notices[0].Severity() != ERROR {
		t.Errorf("Expected ERROR severity, got %s", notices[0].Severity())
	}
}

func TestNoticeContainer_WithLimit(t *testing.T) {
	container := NewNoticeContainerWithLimit(2)

	// Add 3 notices of the same type
	for i := 0; i < 3; i++ {
		notice := NewBaseNotice("test_warning", WARNING, map[string]interface{}{
			"instance": i,
		})
		container.AddNotice(notice)
	}

	notices := container.GetNotices()

	// Should be limited to 2 notices due to maxPerType limit
	if len(notices) != 2 {
		t.Errorf("Expected 2 notices (limited by maxPerType), got %d", len(notices))
	}

	// All notices should have the same code
	for _, notice := range notices {
		if notice.Code() != "test_warning" {
			t.Errorf("Expected notice code 'test_warning', got %s", notice.Code())
		}
	}
}

func TestNoticeContainer_MultipleSeverities(t *testing.T) {
	container := NewNoticeContainer()

	// Add notices of different severities
	container.AddNotice(NewBaseNotice("error1", ERROR, nil))
	container.AddNotice(NewBaseNotice("warning1", WARNING, nil))
	container.AddNotice(NewBaseNotice("info1", INFO, nil))
	container.AddNotice(NewBaseNotice("error2", ERROR, nil))

	notices := container.GetNotices()

	if len(notices) != 4 {
		t.Errorf("Expected 4 notice groups, got %d", len(notices))
	}

	// Count by severity
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, notice := range notices {
		switch notice.Severity() {
		case ERROR:
			errorCount++
		case WARNING:
			warningCount++
		case INFO:
			infoCount++
		}
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 errors, got %d", errorCount)
	}
	if warningCount != 1 {
		t.Errorf("Expected 1 warning, got %d", warningCount)
	}
	if infoCount != 1 {
		t.Errorf("Expected 1 info, got %d", infoCount)
	}
}

func TestNoticeContainer_SameCodeAggregation(t *testing.T) {
	container := NewNoticeContainer()

	// Add multiple notices with the same code
	for i := 0; i < 5; i++ {
		notice := NewBaseNotice("duplicate_field", ERROR, map[string]interface{}{
			"filename":  "routes.txt",
			"rowNumber": i + 2,
			"fieldName": "route_id",
		})
		container.AddNotice(notice)
	}

	notices := container.GetNotices()

	// Should have 5 notices all with the same code
	if len(notices) != 5 {
		t.Errorf("Expected 5 notices, got %d", len(notices))
	}

	for _, notice := range notices {
		if notice.Code() != "duplicate_field" {
			t.Errorf("Expected code 'duplicate_field', got %s", notice.Code())
		}
	}
}

func TestNoticeContainer_ThreadSafety(t *testing.T) {
	container := NewNoticeContainer()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			notice := NewBaseNotice("concurrent_test", WARNING, map[string]interface{}{
				"goroutine": id,
			})
			container.AddNotice(notice)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	notices := container.GetNotices()
	if len(notices) != 10 {
		t.Errorf("Expected 10 notices, got %d", len(notices))
	}

	// All should have the same code
	for _, notice := range notices {
		if notice.Code() != "concurrent_test" {
			t.Errorf("Expected notice code 'concurrent_test', got %s", notice.Code())
		}
	}
}

// Test helper functions for creating common notices
func TestNoticeHelpers(t *testing.T) {
	t.Run("NewMissingRequiredFileNotice", func(t *testing.T) {
		notice := NewMissingRequiredFileNotice("trips.txt")

		if notice.Code() != "missing_required_file" {
			t.Errorf("Expected code 'missing_required_file', got %s", notice.Code())
		}
		if notice.Severity() != ERROR {
			t.Errorf("Expected ERROR severity, got %s", notice.Severity())
		}
		if filename, ok := notice.Context()["filename"].(string); !ok || filename != "trips.txt" {
			t.Errorf("Expected filename 'trips.txt' in context, got %v", notice.Context()["filename"])
		}
	})
}

func TestNoticeContainer_CountBySeverity(t *testing.T) {
	container := NewNoticeContainer()

	// Add mixed notices
	container.AddNotice(NewBaseNotice("error1", ERROR, nil))
	container.AddNotice(NewBaseNotice("error2", ERROR, nil))
	container.AddNotice(NewBaseNotice("error3", ERROR, nil))
	container.AddNotice(NewBaseNotice("warning1", WARNING, nil))
	container.AddNotice(NewBaseNotice("info1", INFO, nil))

	counts := container.CountBySeverity()

	if counts[ERROR] != 3 {
		t.Errorf("Expected error count 3, got %d", counts[ERROR])
	}
	if counts[WARNING] != 1 {
		t.Errorf("Expected warning count 1, got %d", counts[WARNING])
	}
	if counts[INFO] != 1 {
		t.Errorf("Expected info count 1, got %d", counts[INFO])
	}

	totalCount := len(container.GetNotices())
	if totalCount != 5 {
		t.Errorf("Expected total count 5, got %d", totalCount)
	}
}

func TestNotice_Fields(t *testing.T) {
	notice := NewBaseNotice("test_notice", WARNING, map[string]interface{}{
		"filename":  "routes.txt",
		"rowNumber": 5,
	})

	// Test that the notice fields are accessible
	if notice.Code() == "" {
		t.Error("Notice code should not be empty")
	}
	if notice.Code() != "test_notice" {
		t.Errorf("Expected code 'test_notice', got %s", notice.Code())
	}
	if notice.Severity() != WARNING {
		t.Errorf("Expected WARNING severity, got %s", notice.Severity())
	}
	if filename, ok := notice.Context()["filename"].(string); !ok || filename != "routes.txt" {
		t.Errorf("Expected filename 'routes.txt' in context, got %v", notice.Context()["filename"])
	}
}

// TestNoticeContainerKeepsEverythingByDefault guards against a per-type cap
// creeping back in. Capping silently discards notices, and since they arrive
// in file order rather than severity order, the discarded ones can be errors.
func TestNoticeContainerKeepsEverythingByDefault(t *testing.T) {
	container := NewNoticeContainer()

	const total = 5000
	for i := 0; i < total; i++ {
		container.AddNotice(NewBaseNotice("block_trips_with_overlapping_stop_times", ERROR, map[string]interface{}{
			"csvRowNumber": i + 2,
		}))
	}

	if got := len(container.GetNotices()); got != total {
		t.Errorf("expected all %d notices kept, got %d", total, got)
	}
	if got := container.CountBySeverity()[ERROR]; got != total {
		t.Errorf("expected %d errors counted, got %d", total, got)
	}
}

// TestNoticeContainerWithLimitStillCaps checks the opt-in cap still works for
// callers that knowingly want a truncated report.
func TestNoticeContainerWithLimitStillCaps(t *testing.T) {
	container := NewNoticeContainerWithLimit(10)

	for i := 0; i < 50; i++ {
		container.AddNotice(NewBaseNotice("some_code", WARNING, map[string]interface{}{
			"csvRowNumber": i + 2,
		}))
	}

	if got := len(container.GetNotices()); got != 10 {
		t.Errorf("expected the explicit limit of 10 to apply, got %d", got)
	}
}

// TestNoticeContainerDeduplicates covers the case that motivated dedup: two
// validators independently detecting the same finding and both reporting it.
func TestNoticeContainerDeduplicates(t *testing.T) {
	container := NewNoticeContainer()

	overlap := func() Notice {
		return NewBaseNotice("block_trips_with_overlapping_stop_times", ERROR, map[string]interface{}{
			"blockId":        "B1",
			"trip1Id":        "2.112",
			"trip1RowNumber": 910,
			"trip2RowNumber": 917,
		})
	}

	container.AddNotice(overlap())
	container.AddNotice(overlap()) // same finding, reported by a second validator

	if got := len(container.GetNotices()); got != 1 {
		t.Errorf("expected the duplicate to be dropped, got %d notices", got)
	}

	// A different trip is a different finding and must survive.
	container.AddNotice(NewBaseNotice("block_trips_with_overlapping_stop_times", ERROR, map[string]interface{}{
		"blockId":        "B1",
		"trip1Id":        "2.116",
		"trip1RowNumber": 942,
		"trip2RowNumber": 954,
	}))
	if got := len(container.GetNotices()); got != 2 {
		t.Errorf("expected a distinct finding to be kept, got %d notices", got)
	}

	// Same context, different code is also a different finding.
	container.AddNotice(NewBaseNotice("duplicate_stop_in_trip", INFO, map[string]interface{}{
		"tripId":         "2.112",
		"stopId":         "3",
		"firstRowNumber": 910,
		"lastRowNumber":  917,
	}))
	if got := len(container.GetNotices()); got != 3 {
		t.Errorf("expected a distinct code to be kept, got %d notices", got)
	}
}

// TestNoticeIdentityIgnoresMapOrder guards the sort in noticeIdentity: Go
// randomises map iteration, so an unsorted key walk would hash the same
// context differently on different runs and defeat dedup.
func TestNoticeIdentityIgnoresMapOrder(t *testing.T) {
	container := NewNoticeContainer()

	context := map[string]interface{}{
		"a": 1, "b": "two", "c": 3.5, "d": true,
		"e": 5, "f": "six", "g": 7.5, "h": false,
	}

	first := container.noticeIdentity("some_code", context)
	for i := 0; i < 200; i++ {
		if container.noticeIdentity("some_code", context) != first {
			t.Fatal("identity is not stable across map iteration orders")
		}
	}
}
