package goe

type database struct {
	tables map[string]*table
}

func (db *database) Select(...string) From {
	return db
}

func (db *database) From(table string) Table {
	return db.tables[table]
}
