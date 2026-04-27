package theme

import "testing"

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(Theme{Name: "Foo"})
	r.Register(Theme{Name: "Bar"})

	if got, ok := r.Get("foo"); !ok || got.Name != "Foo" {
		t.Fatalf("expected case-insensitive lookup to return Foo, got %q ok=%v", got.Name, ok)
	}
	if _, ok := r.Get("missing"); ok {
		t.Fatal("expected missing theme lookup to fail")
	}
}

func TestRegistryDefaultFallsBackToFirstRegistered(t *testing.T) {
	r := NewRegistry()
	r.Register(Theme{Name: "First"})
	r.Register(Theme{Name: "Second"})

	if got := r.Default(); got.Name != "First" {
		t.Fatalf("expected default to be first registered, got %q", got.Name)
	}
	if err := r.SetDefault("second"); err != nil {
		t.Fatalf("SetDefault returned unexpected error: %v", err)
	}
	if got := r.Default(); got.Name != "Second" {
		t.Fatalf("expected default to be Second after SetDefault, got %q", got.Name)
	}
	if err := r.SetDefault("unknown"); err == nil {
		t.Fatal("expected SetDefault on unknown theme to error")
	}
}

func TestRegistryNextCycles(t *testing.T) {
	r := NewRegistry()
	r.Register(Theme{Name: "A"})
	r.Register(Theme{Name: "B"})
	r.Register(Theme{Name: "C"})

	seq := []string{"A", "B", "C", "A"}
	current := "A"
	for i := 1; i < len(seq); i++ {
		next := r.Next(current)
		if next.Name != seq[i] {
			t.Fatalf("step %d: expected %q, got %q", i, seq[i], next.Name)
		}
		current = next.Name
	}
}

func TestBuiltinHasKnownThemes(t *testing.T) {
	r := Builtin()
	for _, name := range []string{"classic", "mono", "nightowl"} {
		if _, ok := r.Get(name); !ok {
			t.Errorf("expected builtin registry to contain %q", name)
		}
	}
	if got := r.Default().Name; got != "classic" {
		t.Fatalf("expected default theme classic, got %q", got)
	}
}
