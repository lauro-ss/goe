package goe

import "fmt"

type database struct {
	attr   []string
	tables map[string]*table
}

func (db *database) Select(s any, a ...string) From {
	db.attr = a
	return db
}

func (db *database) From(table string) Table {
	for _, v := range db.attr {
		fmt.Println(table + "." + v)
	}
	return db.tables[table]
}
