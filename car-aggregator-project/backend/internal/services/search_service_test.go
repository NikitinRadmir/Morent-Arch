package services

import (
	"testing"

	"car-aggregator/internal/validators"
)

func TestNewSearchServiceDoesNotLoginToCarAPIOnConstruction(t *testing.T) {
	t.Setenv("CARAPI_TOKEN", "invalid-token")
	t.Setenv("CARAPI_SECRET", "invalid-secret")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewSearchService should not panic when CarAPI is unavailable: %v", r)
		}
	}()

	if svc := NewSearchService(nil, validators.NewUserSearchRequestValidator()); svc == nil {
		t.Fatal("expected search service")
	}
}
