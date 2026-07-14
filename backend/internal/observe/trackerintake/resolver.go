package trackerintake

import (
	"fmt"

	"github.com/tinhtran/thanos/backend/internal/domain"
	"github.com/tinhtran/thanos/backend/internal/ports"
)

// MapTrackerResolver routes each configured provider to its own tracker adapter.
// It is the multi-provider counterpart to SingleTrackerResolver: register one
// adapter per domain.TrackerProvider and intake dispatches by the project's
// configured provider. An empty provider resolves to the github default, which
// matches TrackerIntakeConfig.WithDefaults.
type MapTrackerResolver map[domain.TrackerProvider]ports.Tracker

// Resolve returns the adapter registered for provider, or an error when none is
// registered.
func (m MapTrackerResolver) Resolve(provider domain.TrackerProvider) (ports.Tracker, error) {
	if provider == "" {
		provider = domain.TrackerProviderGitHub
	}
	if t, ok := m[provider]; ok && t != nil {
		return t, nil
	}
	return nil, fmt.Errorf("tracker intake: no adapter for provider %q", provider)
}
