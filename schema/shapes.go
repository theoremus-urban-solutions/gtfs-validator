package schema

// Shape represents a shape point from shapes.txt
type Shape struct {
	ShapeID           string   `csv:"shape_id"`
	ShapePtLat        float64  `csv:"shape_pt_lat"`
	ShapePtLon        float64  `csv:"shape_pt_lon"`
	ShapePtSequence   int      `csv:"shape_pt_sequence"`
	ShapeDistTraveled *float64 `csv:"shape_dist_traveled"`
}
