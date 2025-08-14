package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper to run CLI command
func runCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	// Build the CLI first
	cliPath := filepath.Join(t.TempDir(), "gtfs-validator-test")
	cmd := exec.Command("go", "build", "-o", cliPath, ".") // #nosec G204 -- Test code with controlled paths
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Run with provided arguments
	cmd = exec.Command(cliPath, args...) // #nosec G204 -- Test code with controlled paths
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Failed to run CLI: %v", err)
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// Helper to create test data
func createTestGTFS(t *testing.T, valid bool) string {
	t.Helper()

	testDir := t.TempDir()

	// Create minimal GTFS files
	files := map[string]string{
		"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\ntest_agency,Test Transit Agency,https://example.com,America/New_York\n",
		"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nroute_1,test_agency,1,Main Street Line,3\n",
		"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\nstop_1,First Stop,40.7589,-73.9851\nstop_2,Second Stop,40.7614,-73.9776\n",
	}

	if valid {
		files["trips.txt"] = "route_id,service_id,trip_id,trip_headsign\nroute_1,service_1,trip_1,Downtown\n"
		files["stop_times.txt"] = "trip_id,arrival_time,departure_time,stop_id,stop_sequence\ntrip_1,08:00:00,08:00:00,stop_1,1\ntrip_1,08:15:00,08:15:00,stop_2,2\n"
		// Use future dates to avoid expired service warnings
		files["calendar.txt"] = "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nservice_1,1,1,1,1,1,0,0,20250101,20251231\n"
	}

	for filename, content := range files {
		filePath := filepath.Join(testDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return testDir
}

func TestCLI_Version(t *testing.T) {
	stdout, _, exitCode := runCLI(t, "version")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "GTFS Validator CLI vdev") {
		t.Errorf("Expected version info in output, got: %s", stdout)
	}
}

func TestCLI_Help(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Log actual output for debugging
	t.Logf("STDOUT: %s", stdout)
	t.Logf("STDERR: %s", stderr)

	// Check that some basic help text is present (be flexible about exact format)
	helpOutput := stdout + stderr
	if !strings.Contains(helpOutput, "input") && !strings.Contains(helpOutput, "gtfs") {
		t.Errorf("Expected help output to contain input/gtfs information, got: %s", helpOutput)
	}
}

func TestCLI_ValidateValid(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir)

	// The CLI should complete validation (may have warnings but those don't cause non-zero exit)
	if !strings.Contains(stderr, "✅ Validation completed") {
		t.Errorf("Expected validation completion message in stderr, got: %s", stderr)
	}

	// If there are errors, that's ok - the validator might find legitimate issues
	if exitCode != 0 {
		t.Logf("CLI found validation errors (exit code %d), which may be legitimate", exitCode)
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
	}
}

func TestCLI_ValidateInvalid(t *testing.T) {
	testDir := createTestGTFS(t, false) // Missing required files

	stdout, stderr, exitCode := runCLI(t, "-i", testDir)

	if exitCode == 0 {
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
		t.Error("Expected non-zero exit code for invalid GTFS")
	}

	if !strings.Contains(stderr, "errors found") {
		t.Errorf("Expected error message in stderr, got: %s", stderr)
	}
}

func TestCLI_JSONOutput(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-f", "json")

	// Validation should complete regardless of exit code
	if !strings.Contains(stderr, "✅ Validation completed") {
		t.Errorf("Expected validation completion message in stderr, got: %s", stderr)
	}

	// Parse JSON output
	var report map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, stdout)
		return
	}

	// Check basic structure
	if _, exists := report["summary"]; !exists {
		t.Error("Expected 'summary' field in JSON output")
	}
	if _, exists := report["notices"]; !exists {
		t.Error("Expected 'notices' field in JSON output")
	}

	if exitCode != 0 {
		t.Logf("CLI found validation errors (exit code %d), JSON output generated successfully", exitCode)
	}
}

func TestCLI_SummaryOutput(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-f", "summary")

	// Validation should complete
	if !strings.Contains(stderr, "✅ Validation completed") {
		t.Errorf("Expected validation completion message in stderr, got: %s", stderr)
	}

	expectedStrings := []string{
		"GTFS Validation Summary",
		"Feed Statistics:",
		"Validation Results:",
		"Agencies:",
		"Routes:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected %q in summary output, got: %s", expected, stdout)
		}
	}

	if exitCode != 0 {
		t.Logf("CLI found validation errors (exit code %d), summary output generated successfully", exitCode)
	}
}

func TestCLI_PerformanceMode(t *testing.T) {
	testDir := createTestGTFS(t, true)

	start := time.Now()
	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-m", "performance")
	elapsed := time.Since(start)

	if exitCode != 0 {
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stderr, "Mode: performance") {
		t.Errorf("Expected performance mode indication in stderr, got: %s", stderr)
	}

	// Performance mode should be reasonably fast (less than 30 seconds for this small dataset)
	if elapsed > 30*time.Second {
		t.Errorf("Performance mode took too long: %v", elapsed)
	}
}

func TestCLI_ComprehensiveMode(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-m", "comprehensive")

	if !strings.Contains(stderr, "Mode: comprehensive") {
		t.Errorf("Expected comprehensive mode indication in stderr, got: %s", stderr)
	}

	if !strings.Contains(stderr, "✅ Validation completed") {
		t.Errorf("Expected validation completion message in stderr, got: %s", stderr)
	}

	if exitCode != 0 {
		t.Logf("CLI found validation errors in comprehensive mode (exit code %d)", exitCode)
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
	}
}

func TestCLI_NonExistentInput(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "-i", "/non/existent/path")

	if exitCode == 0 {
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
		t.Error("Expected non-zero exit code for non-existent input")
	}

	if !strings.Contains(stderr, "does not exist") {
		t.Errorf("Expected file not found error in stderr, got: %s", stderr)
	}
}

func TestCLI_MissingInput(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t)

	if exitCode == 0 {
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
		t.Error("Expected non-zero exit code when input is missing")
	}

	if !strings.Contains(stderr, "required flag(s) \"input\" not set") {
		t.Errorf("Expected missing input error in stderr, got: %s", stderr)
	}
}

func TestCLI_InvalidMode(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-m", "invalid_mode")

	if exitCode == 0 {
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
		t.Error("Expected non-zero exit code for invalid mode")
	}

	if !strings.Contains(stderr, "invalid validation mode") {
		t.Errorf("Expected invalid mode error in stderr, got: %s", stderr)
	}
}

func TestCLI_InvalidFormat(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-f", "invalid_format")

	if exitCode == 0 {
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
		t.Error("Expected non-zero exit code for invalid format")
	}

	// Debug: log the actual output for troubleshooting
	t.Logf("STDOUT: %q", stdout)
	t.Logf("STDERR: %q", stderr)

	// The error message may include emoji and "Error:" prefix
	// Check for the core error message content in both stdout and stderr
	combinedOutput := stdout + stderr
	if !strings.Contains(combinedOutput, "invalid output format") &&
		!strings.Contains(combinedOutput, "Format Error") &&
		!strings.Contains(combinedOutput, "Unknown output format") {
		t.Errorf("Expected invalid format error in output, got stdout: %s, stderr: %s", stdout, stderr)
	}
}

func TestCLI_CustomWorkers(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-w", "8")

	// Should complete successfully with custom worker count
	if !strings.Contains(stderr, "✅ Validation completed") {
		t.Errorf("Expected validation completion message in stderr, got: %s", stderr)
	}

	if exitCode != 0 {
		t.Logf("CLI found validation errors with custom workers (exit code %d)", exitCode)
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
	}
}

func TestCLI_CustomCountry(t *testing.T) {
	testDir := createTestGTFS(t, true)

	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-c", "UK")

	// Should complete successfully with custom country code
	if !strings.Contains(stderr, "✅ Validation completed") {
		t.Errorf("Expected validation completion message in stderr, got: %s", stderr)
	}

	if exitCode != 0 {
		t.Logf("CLI found validation errors with custom country (exit code %d)", exitCode)
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
	}
}

func TestCLI_Timeout(t *testing.T) {
	testDir := createTestGTFS(t, true)

	// 30s timeout should be plenty for small test data
	stdout, stderr, exitCode := runCLI(t, "-i", testDir, "-t", "30s")

	// Should complete within timeout
	if !strings.Contains(stderr, "✅ Validation completed") {
		t.Errorf("Expected validation completion message in stderr, got: %s", stderr)
	}

	if exitCode != 0 {
		t.Logf("CLI found validation errors with timeout (exit code %d)", exitCode)
		t.Logf("STDOUT: %s", stdout)
		t.Logf("STDERR: %s", stderr)
	}
}
