package agents

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers []Provider) Registry {
	registry := Registry{providers: map[string]Provider{}}
	for _, provider := range providers {
		registry.providers[provider.ID] = provider
	}
	return registry
}

func (r Registry) Get(id string) (Provider, bool) {
	provider, ok := r.providers[id]
	return provider, ok
}
