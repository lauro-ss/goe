package goe

import (
	"reflect"
)

// func Connect(u string, c Config) *database {
// 	return &database{tables: make(map[string]*table)}
// }

func Connect(db any) {
	value := reflect.ValueOf(db).Elem()
	value.FieldByName("Database").Set(reflect.ValueOf(&database{}))
}
