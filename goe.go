package goe

import (
	"reflect"
)

// func Connect(u string, c Config) *database {
// 	return &database{tables: make(map[string]*table)}
// }

func Connect(db any) {
	value := reflect.ValueOf(db).Elem()
	value.FieldByName("DB").Set(reflect.ValueOf(&DB{}))

	// for i := 0; i < value.NumField(); i++ {
	// 	fmt.Println(value.Field(i).Elem().Type())
	// }
	//fmt.Println(value.Field(0).Elem().Type().Name())
}
func Map(t any, s any) {
	tab := reflect.TypeOf(t).Elem()
	str := reflect.TypeOf(s)
	value := reflect.ValueOf(t).Elem()

	var field reflect.StructField
	for i := 0; i < tab.NumField(); i++ {
		field, _ = str.FieldByName(tab.Field(i).Name)
		value = value.FieldByName(tab.Field(i).Name)
		mapAttribute(value, tab.Field(i), field, str.Name())
	}
}

func mapAttribute(v reflect.Value, t reflect.StructField, s reflect.StructField, tableName string) {
	at := &att{}
	at.name = s.Name
	at.table = tableName

	v.Set(reflect.ValueOf(at))
}
