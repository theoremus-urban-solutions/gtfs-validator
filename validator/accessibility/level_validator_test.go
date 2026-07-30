package accessibility

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestLevelValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		expected    map[string]int
		description string
	}{
		{
			name: "levels used by stops",
			files: map[string]string{
				"levels.txt": "level_id,level_index,level_name\nL1,0,Ground\nL2,1,Mezzanine",
				"stops.txt":  "stop_id,stop_name,level_id\nS1,Stop 1,L1\nS2,Stop 2,L2",
			},
			expected:    map[string]int{},
			description: "Every level is referenced and each index is distinct",
		},
		{
			name: "level nobody references",
			files: map[string]string{
				"levels.txt": "level_id,level_index,level_name\nL1,0,Ground\nL2,1,Mezzanine",
				"stops.txt":  "stop_id,stop_name,level_id\nS1,Stop 1,L1",
			},
			expected:    map[string]int{"unused_level": 1},
			description: "L2 is defined but no stop or pathway stands on it",
		},
		{
			name: "two levels claiming the same index",
			files: map[string]string{
				"levels.txt": "level_id,level_index,level_name\nL1,0,Ground\nL2,0,Also ground",
				"stops.txt":  "stop_id,stop_name,level_id\nS1,Stop 1,L1\nS2,Stop 2,L2",
			},
			expected:    map[string]int{"duplicate_level_index": 1},
			description: "level_index orders the levels, so two levels cannot share one",
		},
		{
			// level_name is Optional in the spec, not Recommended. Canonical
			// reports nothing for an unnamed level, and a station's service
			// levels are routinely unnamed.
			name: "level with no name",
			files: map[string]string{
				"levels.txt": "level_id,level_index,level_name\nL1,0,\nL2,-8,",
				"stops.txt":  "stop_id,stop_name,level_id\nS1,Stop 1,L1\nS2,Stop 2,L2",
			},
			expected:    map[string]int{},
			description: "An unnamed level is not a defect",
		},
		{
			name:        "no levels file",
			files:       map[string]string{"stops.txt": "stop_id,stop_name\nS1,Stop 1"},
			expected:    map[string]int{},
			description: "levels.txt is optional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			NewLevelValidator().Validate(loader, container, gtfsvalidator.Config{})

			codes := map[string]int{}
			for _, n := range container.GetNotices() {
				codes[n.Code()]++
			}

			for _, code := range []string{"unused_level", "duplicate_level_index", "missing_recommended_field"} {
				if _, named := tt.expected[code]; !named {
					tt.expected[code] = 0
				}
			}
			for code, want := range tt.expected {
				if codes[code] != want {
					t.Errorf("%s: got %d, want %d — %s", code, codes[code], want, tt.description)
				}
			}
		})
	}
}

func TestLevelValidator_New(t *testing.T) {
	if NewLevelValidator() == nil {
		t.Error("NewLevelValidator() returned nil")
	}
}
