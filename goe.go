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
	return mapData(database, database.FieldByName(str.Name()), str)
}

// Map makes the mapping from the struct fields "s" to the target "t"
func mapData(database reflect.Value, target reflect.Value, str reflect.Type) error {
	field, exists := str.FieldByName("Id")
	var pk *Pk
	if exists {
		pk = mapPrimaryKey(target.FieldByName("Id"), field, str.Name())
	}

	for i := 0; i < str.NumField(); i++ {
		field, exists = target.Type().FieldByName(str.Field(i).Name)
		if exists {
			switch target.Field(i).Interface().(type) {
			case *Att:
				mapAttribute(target.Field(i), field, pk)
			}
		} else {
			checkField(database, pk, str.Field(i))
		}
	}

	return nil
}

func mapPrimaryKey(target reflect.Value, field reflect.StructField, tableName string) *Pk {
	if !target.IsNil() {
		return (*Pk)(target.UnsafePointer())
	}
	p := &Pk{name: field.Name, table: tableName, Fk: make(map[string]*Pk)}
	target.Set(reflect.ValueOf(p))
	return (*Pk)(target.UnsafePointer())
}

func mapAttribute(target reflect.Value, field reflect.StructField, pk *Pk) {
	at := &Att{pk: pk}
	at.name = field.Name

	target.Set(reflect.ValueOf(at))
}

func checkField(database reflect.Value, pk *Pk, field reflect.StructField) {
	switch field.Type.Kind() {
	case reflect.Struct:
		//many to one
		fmt.Println("Struct")
	case reflect.Slice:
		//possibile many to many
		str := field.Type.Elem()
		target := database.FieldByName(str.Name()).FieldByName("Id")
		fmt.Println(str, pk)
		if target.IsZero() {
			field, _ := str.FieldByName("Id")
			pk.Fk[str.Name()] = mapPrimaryKey(target, field, str.Name())
		} else {
			pk.Fk[str.Name()] = getPrimaryKey(database, str)
		}
	}

}

func getPrimaryKey(database reflect.Value, str reflect.Type) *Pk {
	field := database.FieldByName(str.Name()).FieldByName("Id")
	return (*Pk)(field.UnsafePointer())
}
