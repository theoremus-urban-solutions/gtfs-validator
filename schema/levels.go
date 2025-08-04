package schema

// Level represents level information from levels.txt
type Level struct {
	LevelID    string  `csv:"level_id"`
	LevelIndex float64 `csv:"level_index"`
	LevelName  string  `csv:"level_name"`
}