package theme

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Registry struct {
	mu         sync.RWMutex
	themes     map[string]Theme
	order      []string
	defaultKey string
}

func NewRegistry() *Registry {
	return &Registry{themes: make(map[string]Theme)}
}

func (r *Registry) Register(t Theme) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := normalize(t.Name)
	if _, exists := r.themes[key]; !exists {
		r.order = append(r.order, key)
	}
	r.themes[key] = t
	if r.defaultKey == "" {
		r.defaultKey = key
	}
}

func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := normalize(name)
	if _, ok := r.themes[key]; !ok {
		return fmt.Errorf("theme %q is not registered", name)
	}
	r.defaultKey = key
	return nil
}

func (r *Registry) Get(name string) (Theme, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.themes[normalize(name)]
	return t, ok
}

func (r *Registry) Default() Theme {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.themes[r.defaultKey]; ok {
		return t
	}
	for _, key := range r.order {
		return r.themes[key]
	}
	return Theme{}
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	sort.Strings(out)
	return out
}

func (r *Registry) Next(current string) Theme {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.order) == 0 {
		return Theme{}
	}
	key := normalize(current)
	for i, name := range r.order {
		if name == key {
			return r.themes[r.order[(i+1)%len(r.order)]]
		}
	}
	return r.themes[r.order[0]]
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
