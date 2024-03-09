package goe

import (
	"fmt"
	"reflect"
)

// func Connect(u string, c Config) *database {
// 	return &database{tables: make(map[string]*table)}
// }

// func Connect(db any) {
// 	value := reflect.ValueOf(db).Elem()
// 	value.FieldByName("DB").Set(reflect.ValueOf(&DB{}))

// 	// for i := 0; i < value.NumField(); i++ {
// 	// 	fmt.Println(value.Field(i).Elem().Type())
// 	// }
// 	//fmt.Println(value.Field(0).Elem().Type().Name())
// }

func Map(db any, s any) error {
	if reflect.ValueOf(db).Kind() != reflect.Ptr {
		return fmt.Errorf("%v: the target value needs to be pass as a pointer", pkg)
	}
	database := reflect.ValueOf(db).Elem()
	str := reflect.TypeOf(s)
	return mapData(database.FieldByName(str.Name()), str)
}

// Map makes the mapping from the struct fields "s" to the target "t"
func mapData(target reflect.Value, str reflect.Type) error {
	field, exists := str.FieldByName("Id")
	var pk *Pk
	if exists {
		pk = mapPrimaryKey(target.FieldByName("Id"), field, str.Name())
	}
	fmt.Printf("%p Aquii\n", pk)

	for i := 0; i < str.NumField(); i++ {
		field, exists = target.Type().FieldByName(str.Field(i).Name)
		if exists {
			switch target.Field(i).Interface().(type) {
			case Att:
				mapAttribute(target.Field(i), field)
			}
		} else {
			checkField(target.FieldByName("Id"), str.Field(i))
		}
	}

	return nil
}

func mapPrimaryKey(target reflect.Value, field reflect.StructField, tableName string) *Pk {
	pk := &Pk{name: field.Name, table: tableName}
	target.Set(reflect.ValueOf(pk))
	return (*Pk)(target.UnsafePointer())
}

func mapAttribute(target reflect.Value, field reflect.StructField) {
	at := Att{}
	at.name = field.Name

	target.Set(reflect.ValueOf(at))
}

func checkField(target reflect.Value, f reflect.StructField) {
	fmt.Println(f, target)
}
