package goe

func Connect() Database {
	return &database{tables: make(map[string]*table)}
}
