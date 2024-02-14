package goe

func Connect(u string, c Config) *database {
	return &database{tables: make(map[string]*table)}
}
