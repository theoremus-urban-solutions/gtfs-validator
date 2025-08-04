package accessibility

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// LevelValidator validates level definitions for multi-level stations
type LevelValidator struct{}

// NewLevelValidator creates a new level validator
func NewLevelValidator() *LevelValidator {
	return &LevelValidator{}
}

// LevelInfo represents level information
type LevelInfo struct {
	LevelID    string
	LevelIndex float64
	LevelName  string
	RowNumber  int
}

// Validate checks level definitions
func (v *LevelValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	levels := v.loadLevels(loader)
	if len(levels) == 0 {
		return // No levels to validate
	}

	// Validate each level
	for _, level := range levels {
		v.validateLevel(container, level)
	}

	// Check for duplicate level indices
	v.validateDuplicateLevelIndices(container, levels)

	// Validate level usage
	v.validateLevelUsage(loader, container, levels)
}

// loadLevels loads level information from levels.txt
func (v *LevelValidator) loadLevels(loader *parser.FeedLoader) map[string]*LevelInfo {
	levels := make(map[string]*LevelInfo)

	reader, err := loader.GetFile("levels.txt")
	if err != nil {
		return levels
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "levels.txt")
	if err != nil {
		return levels
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		level := v.parseLevel(row)
		if level != nil {
			levels[level.LevelID] = level
		}
	}

	return levels
}

// parseLevel parses a level record
func (v *LevelValidator) parseLevel(row *parser.CSVRow) *LevelInfo {
	levelID, hasLevelID := row.Values["level_id"]
	levelIndexStr, hasLevelIndex := row.Values["level_index"]

	if !hasLevelID || !hasLevelIndex {
		return nil
	}

	levelIndex, err := strconv.ParseFloat(strings.TrimSpace(levelIndexStr), 64)
	if err != nil {
		return nil
	}

	level := &LevelInfo{
		LevelID:    strings.TrimSpace(levelID),
		LevelIndex: levelIndex,
		RowNumber:  row.RowNumber,
	}

	// Parse optional level name
	if levelName, hasLevelName := row.Values["level_name"]; hasLevelName {
		level.LevelName = strings.TrimSpace(levelName)
	}

	return level
}

// validateLevel validates a single level record
func (v *LevelValidator) validateLevel(container *notice.NoticeContainer, level *LevelInfo) {
	// Validate level index range (reasonable bounds)
	if level.LevelIndex < -50 || level.LevelIndex > 50 {
		container.AddNotice(notice.NewUnreasonableLevelIndexNotice(
			level.LevelID,
			level.LevelIndex,
			level.RowNumber,
		))
	}

	// Check for missing level name (recommended)
	if level.LevelName == "" {
		container.AddNotice(notice.NewMissingRecommendedFieldNotice(
			"levels.txt",
			"level_name",
			level.RowNumber,
		))
	}
}

// validateDuplicateLevelIndices checks for duplicate level indices
func (v *LevelValidator) validateDuplicateLevelIndices(container *notice.NoticeContainer, levels map[string]*LevelInfo) {
	indexMap := make(map[float64][]*LevelInfo)

	for _, level := range levels {
		indexMap[level.LevelIndex] = append(indexMap[level.LevelIndex], level)
	}

	for levelIndex, levelList := range indexMap {
		if len(levelList) > 1 {
			for i := 1; i < len(levelList); i++ {
				container.AddNotice(notice.NewDuplicateLevelIndexNotice(
					levelList[i].LevelID,
					levelIndex,
					levelList[i].RowNumber,
					levelList[0].RowNumber,
				))
			}
		}
	}
}

// validateLevelUsage checks if levels are actually used by stops or pathways
func (v *LevelValidator) validateLevelUsage(loader *parser.FeedLoader, container *notice.NoticeContainer, levels map[string]*LevelInfo) {
	// Check usage in stops.txt
	usedLevels := v.loadUsedLevelsFromStops(loader)

	// Check usage in pathways.txt
	usedInPathways := v.loadUsedLevelsFromPathways(loader)

	// Combine usage
	for levelID := range usedInPathways {
		usedLevels[levelID] = true
	}

	// Check for unused levels
	for levelID, level := range levels {
		if !usedLevels[levelID] {
			container.AddNotice(notice.NewUnusedLevelNotice(
				levelID,
				level.RowNumber,
			))
		}
	}
}

// loadUsedLevelsFromStops loads level IDs used in stops.txt
func (v *LevelValidator) loadUsedLevelsFromStops(loader *parser.FeedLoader) map[string]bool {
	usedLevels := make(map[string]bool)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return usedLevels
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return usedLevels
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if levelID, hasLevelID := row.Values["level_id"]; hasLevelID && strings.TrimSpace(levelID) != "" {
			usedLevels[strings.TrimSpace(levelID)] = true
		}
	}

	return usedLevels
}

// loadUsedLevelsFromPathways loads level IDs used in pathways.txt
func (v *LevelValidator) loadUsedLevelsFromPathways(loader *parser.FeedLoader) map[string]bool {
	usedLevels := make(map[string]bool)

	reader, err := loader.GetFile("pathways.txt")
	if err != nil {
		return usedLevels
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "pathways.txt")
	if err != nil {
		return usedLevels
	}

	for {
		_, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		// Check from_stop_id and to_stop_id level references
		// Note: This is a simplified check - in reality we'd need to check
		// if the referenced stops have level_id fields
	}

	return usedLevels
}
