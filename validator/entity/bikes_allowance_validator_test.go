package entity

import (
	"testing"
)

func TestBikesAllowanceValidator_IsBikeUncommonRouteType(t *testing.T) {
	validator := NewBikesAllowanceValidator()

	uncommonTypes := []int{0, 1, 2, 5, 6, 7} // Tram, Subway, Rail, Cable tram, Aerial lift, Funicular
	commonTypes := []int{3, 4, 11, 12}       // Bus, Ferry, Trolleybus, Monorail

	for _, routeType := range uncommonTypes {
		if !validator.isBikeUncommonRouteType(routeType) {
			t.Errorf("Expected route type %d to be uncommon for bikes", routeType)
		}
	}

	for _, routeType := range commonTypes {
		if validator.isBikeUncommonRouteType(routeType) {
			t.Errorf("Expected route type %d to be common for bikes", routeType)
		}
	}
}

func TestBikesAllowanceValidator_New(t *testing.T) {
	validator := NewBikesAllowanceValidator()
	if validator == nil {
		t.Error("NewBikesAllowanceValidator() returned nil")
	}
}
