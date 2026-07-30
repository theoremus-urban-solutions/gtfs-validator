package relationship

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TripShapeDistanceValidator reports trips whose stop times measure distance
// along a shape that does not measure it on every point. Distances are only
// useful as a pair — one to locate the stop, one to locate the geometry — so a
// trip carrying only half of them cannot have its stops placed on its shape.
type TripShapeDistanceValidator struct{}

// NewTripShapeDistanceValidator creates a new trip shape distance validator
func NewTripShapeDistanceValidator() *TripShapeDistanceValidator {
	return &TripShapeDistanceValidator{}
}

// shapeDistanceCandidate is a trip whose shape lacks distances, held until the
// stop_times pass says whether the trip supplies any of its own.
type shapeDistanceCandidate struct {
	ShapeID   string
	RowNumber int
}

// Validate reports trips with stop time distances against a shape without them.
func (v *TripShapeDistanceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	incomplete := v.loadIncompleteShapes(loader)
	if len(incomplete) == 0 {
		return
	}

	candidates := v.loadCandidateTrips(loader, incomplete)
	if len(candidates) == 0 {
		return
	}

	v.reportTripsWithDistances(loader, container, candidates)
}

// loadIncompleteShapes returns the shapes that leave shape_dist_traveled empty
// on at least one of their points.
func (v *TripShapeDistanceValidator) loadIncompleteShapes(loader *parser.FeedLoader) map[string]bool {
	incomplete := make(map[string]bool)

	reader, err := loader.GetFile("shapes.txt")
	if err != nil {
		return incomplete
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "shapes.txt")
	if err != nil {
		return incomplete
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		shapeID := strings.TrimSpace(row.Values["shape_id"])
		if shapeID == "" {
			continue
		}
		if strings.TrimSpace(row.Values["shape_dist_traveled"]) == "" {
			incomplete[shapeID] = true
		}
	}

	return incomplete
}

// loadCandidateTrips returns the trips that run on a shape without complete
// distances. Trips on a complete shape, or on none at all, are dropped here so
// the stop_times pass has the smallest possible set to test against.
func (v *TripShapeDistanceValidator) loadCandidateTrips(loader *parser.FeedLoader, incomplete map[string]bool) map[string]*shapeDistanceCandidate {
	candidates := make(map[string]*shapeDistanceCandidate)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return candidates
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return candidates
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		tripID := strings.TrimSpace(row.Values["trip_id"])
		shapeID := strings.TrimSpace(row.Values["shape_id"])
		if tripID == "" || !incomplete[shapeID] {
			continue
		}
		candidates[tripID] = &shapeDistanceCandidate{
			ShapeID:   shapeID,
			RowNumber: row.RowNumber,
		}
	}

	return candidates
}

// reportTripsWithDistances streams stop_times.txt and reports each candidate
// trip the first time it supplies a distance. The trip is dropped once
// reported, so the rest of its stop times cost only a map lookup.
func (v *TripShapeDistanceValidator) reportTripsWithDistances(loader *parser.FeedLoader, container *notice.NoticeContainer, candidates map[string]*shapeDistanceCandidate) {
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			return
		}
		if err != nil {
			continue
		}
		if strings.TrimSpace(row.Values["shape_dist_traveled"]) == "" {
			continue
		}

		tripID := strings.TrimSpace(row.Values["trip_id"])
		candidate, wanted := candidates[tripID]
		if !wanted {
			continue
		}
		container.AddNotice(notice.NewTripWithShapeDistTraveledButNoShapeDistancesNotice(
			tripID,
			candidate.ShapeID,
			candidate.RowNumber,
		))
		delete(candidates, tripID)

		if len(candidates) == 0 {
			return
		}
	}
}
