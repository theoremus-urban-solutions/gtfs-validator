// Example of using the GTFS validator in a web API server
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

// ValidationRequest represents an API validation request
type ValidationRequest struct {
	Mode              string `json:"mode,omitempty"`
	CountryCode       string `json:"countryCode,omitempty"`
	MaxNoticesPerType int    `json:"maxNoticesPerType,omitempty"`
}

// ValidationResponse represents the API response
type ValidationResponse struct {
	Success bool                            `json:"success"`
	Report  *gtfsvalidator.ValidationReport `json:"report,omitempty"`
	Error   string                          `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/validate", validateHandler)
	http.HandleFunc("/health", healthHandler)

	port := ":8080"
	fmt.Printf("GTFS Validator API Server starting on %s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /validate - Upload and validate a GTFS file")
	fmt.Println("  GET  /health  - Health check endpoint")

	log.Fatal(http.ListenAndServe(port, nil))
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 100MB)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file
	file, header, err := r.FormFile("file")
	if err != nil {
		sendErrorResponse(w, "Missing 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse optional parameters
	var req ValidationRequest
	if configJSON := r.FormValue("config"); configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &req); err != nil {
			sendErrorResponse(w, "Invalid config JSON", http.StatusBadRequest)
			return
		}
	}

	// Log request
	log.Printf("Validating file: %s (size: %d bytes)", header.Filename, header.Size)

	// Create validator with configuration
	opts := []gtfsvalidator.Option{}

	// Set validation mode
	switch req.Mode {
	case "performance":
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance))
	case "comprehensive":
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeComprehensive))
	default:
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeDefault))
	}

	// Set country code
	if req.CountryCode != "" {
		opts = append(opts, gtfsvalidator.WithCountryCode(req.CountryCode))
	}

	// Set notice limit
	if req.MaxNoticesPerType > 0 {
		opts = append(opts, gtfsvalidator.WithMaxNoticesPerType(req.MaxNoticesPerType))
	}

	validator := gtfsvalidator.New(opts...)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Validate the file
	startTime := time.Now()
	report, err := validator.ValidateReaderWithContext(ctx, file)
	elapsed := time.Since(startTime)

	if err != nil {
		if err == context.DeadlineExceeded {
			sendErrorResponse(w, "Validation timeout exceeded", http.StatusRequestTimeout)
		} else {
			sendErrorResponse(w, fmt.Sprintf("Validation failed: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Log results
	log.Printf("Validation completed in %.2fs - Errors: %d, Warnings: %d",
		elapsed.Seconds(),
		report.ErrorCount(),
		report.WarningCount())

	// Send response
	sendSuccessResponse(w, report)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ValidationResponse{
		Success: false,
		Error:   message,
	})
}

func sendSuccessResponse(w http.ResponseWriter, report *gtfsvalidator.ValidationReport) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ValidationResponse{
		Success: true,
		Report:  report,
	})
}

// Example usage with curl:
// curl -X POST -F "file=@transit-feed.zip" http://localhost:8080/validate
//
// With options:
// curl -X POST \
//   -F "file=@transit-feed.zip" \
//   -F 'config={"mode":"performance","countryCode":"UK","maxNoticesPerType":50}' \
//   http://localhost:8080/validate
