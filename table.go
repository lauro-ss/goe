package goe

type table struct{}

func (t *table) Join(string) Join {
	return t
}
