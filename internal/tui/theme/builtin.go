package theme

func Builtin() *Registry {
	r := NewRegistry()
	r.Register(Classic())
	r.Register(Mono())
	r.Register(NightOwl())
	_ = r.SetDefault("classic")
	return r
}
