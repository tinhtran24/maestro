package trackerintake

import (
	"testing"

	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/ports"
)

func TestMapTrackerResolver(t *testing.T) {
	gh := &fakeTracker{}
	resolver := MapTrackerResolver{domain.TrackerProviderGitHub: gh}

	// Registered provider resolves to its adapter.
	got, err := resolver.Resolve(domain.TrackerProviderGitHub)
	if err != nil {
		t.Fatalf("Resolve(github) err: %v", err)
	}
	if got != ports.Tracker(gh) {
		t.Fatalf("Resolve(github) returned the wrong adapter")
	}

	// Empty provider falls back to the github default.
	if got, err := resolver.Resolve(""); err != nil || got != ports.Tracker(gh) {
		t.Fatalf("Resolve(\"\") = (%v, %v), want the github adapter", got, err)
	}

	// Unregistered provider errors.
	if _, err := resolver.Resolve("gitlab"); err == nil {
		t.Fatal("Resolve(gitlab): want error, got nil")
	}

	// A nil-valued registration is treated as unregistered.
	nilResolver := MapTrackerResolver{domain.TrackerProviderGitHub: nil}
	if _, err := nilResolver.Resolve(domain.TrackerProviderGitHub); err == nil {
		t.Fatal("Resolve(github) with nil adapter: want error, got nil")
	}
}
