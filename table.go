package goe

type table struct {
	name       string
	attributes map[string]*attribute
}

type attribute struct {
	name string
}

func (t *table) Join(string) Join {
	return t
}
