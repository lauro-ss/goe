package goe

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func Map(db any, s any) error {
	if reflect.ValueOf(db).Kind() != reflect.Ptr {
		return fmt.Errorf("%v: the target value needs to be pass as a pointer", pkg)
	}
	database := reflect.ValueOf(db).Elem()
	str := reflect.TypeOf(s)
	target := database.FieldByName(str.Name())
	if target.Kind() == reflect.Invalid {
		log.Printf("goe: target %v is not declared on %v", str.Name(), database.Type().Name())
		return nil
	}
	mapData(database, target, str)
	checkMapping(target, str)
	return nil
}

// Map makes the mapping from the struct fields "s" to the target "t"
func mapData(database reflect.Value, target reflect.Value, str reflect.Type) {
	pks := primaryKeys(str)

	var pk *pk
	if len(pks) > 0 {
		//TODO: Add more then one primary key
		field := pks[0]
		pk = mapPrimaryKey(target.FieldByName(field.Name), field, str.Name())
	} else {
		//TODO: add a anonymous pk for targets without
	}

	attrs := attributes(str, str.NumField()-len(pks), pks)
	for _, attr := range attrs {
		targetField := target.FieldByName(attr.Name)
		if targetField.Kind() != reflect.Invalid {
			mapAttribute(targetField, attr, pk)
		} else {
			if attr.Type.Kind() != reflect.Slice && attr.Type.Kind() != reflect.Struct {
				log.Printf("goe: target %v don't have the attribute \"%v\" for %v", target.Type(), attr.Name, str)
			}
			if pk != nil {
				checkForeignKey(database, pk, attr)
			}
		}
	}
}

func mapPrimaryKey(target reflect.Value, field reflect.StructField, tableName string) *pk {
	if target.Kind() == reflect.Invalid {
		return nil
	}

	if !target.IsNil() {
		return (*pk)(target.Elem().UnsafePointer())
	}
	p := &pk{name: field.Name, table: tableName, Fk: make(map[string]*pk)}
	target.Set(reflect.ValueOf(p))
	return (*pk)(target.Elem().UnsafePointer())
}

func mapAttribute(target reflect.Value, field reflect.StructField, pk *pk) {
	at := &att{pk: pk}
	at.name = field.Name

	target.Set(reflect.ValueOf(at))
}

func checkForeignKey(database reflect.Value, pk *pk, field reflect.StructField) {
	switch field.Type.Kind() {
	case reflect.Struct:
		//possible many to one
		str := field.Type
		for i := 0; i < str.NumField(); i++ {
			switch str.Field(i).Type.Kind() {
			case reflect.Slice:
				if str.Field(i).Type.Elem().Name() == pk.table {
					mapManyToOny(database, pk, str)
				}
			case reflect.Struct:
				fmt.Printf("one %v to one %v \n", str.Name(), pk.table)
			}
		}
	case reflect.Slice:
		//possibile many to many

		str := field.Type.Elem()
		for i := 0; i < str.NumField(); i++ {
			switch str.Field(i).Type.Kind() {
			case reflect.Slice:
				if str.Field(i).Type.Elem().Name() == pk.table {
					mapManyToMany(database, pk, str)
				}
			case reflect.Struct:
				if str.Field(i).Type.Name() == pk.table {
					mapManyToOny(database, pk, str)
				}
			}

		}
	}

}

func mapManyToOny(database reflect.Value, pk *pk, str reflect.Type) {
	key := str.Name()

	//TODO: Add more then one primary key
	target := database.FieldByName(key).FieldByName(primaryKeys(str)[0].Name)
	if target.Kind() == reflect.Invalid {
		return
	}

	if target.IsZero() {
		field := primaryKeys(str)[0]
		pk.Fk[key] = mapPrimaryKey(target, field, str.Name())
		return
	}
	pk.Fk[key] = getPrimaryKey(database, str)
}

func mapManyToMany(database reflect.Value, primary *pk, str reflect.Type) {
	key := str.Name()
	target := database.FieldByName(key + primary.table)
	var table string
	table = key + primary.table
	if target.Kind() == reflect.Invalid {
		target = database.FieldByName(primary.table + key)
		table = primary.table + key
	}

	if target.Kind() == reflect.Invalid {
		//No many to many default target
		return
	}

	//Id + current Target
	pk0 := target.FieldByName("Id" + primary.table)

	if !pk0.IsNil() {
		return
	}

	//Id + target map
	pk1 := target.FieldByName("Id" + key)

	//TODO: Add more then one primary key
	target = database.FieldByName(key).FieldByName(primaryKeys(str)[0].Name)
	if target.Kind() == reflect.Invalid {
		return
	}

	//Fills the target primary key
	field := primaryKeys(str)[0]
	primaryKeyTarget := mapPrimaryKey(target, field, str.Name())

	p := &pk{name: "Id" + primary.table, table: table, Fk: make(map[string]*pk)}
	p.Fk[key] = primaryKeyTarget
	pk0.Set(reflect.ValueOf(p))

	p = &pk{name: "Id" + key, table: table, Fk: make(map[string]*pk)}
	p.Fk[primary.table] = primary
	pk1.Set(reflect.ValueOf(p))

	//default structs points to the many to many target
	primary.Fk[primary.table] = (*pk)(pk0.Elem().UnsafePointer())
	primaryKeyTarget.Fk[key] = (*pk)(pk1.Elem().UnsafePointer())
}

func getPrimaryKey(database reflect.Value, str reflect.Type) *pk {
	//TODO: Add more then one primary key
	field := database.FieldByName(str.Name()).FieldByName(primaryKeys(str)[0].Name)
	return (*pk)(field.Elem().UnsafePointer())
}

func primaryKeys(str reflect.Type) (pks []reflect.StructField) {
	field, exists := str.FieldByName("Id")
	if exists {
		pks := make([]reflect.StructField, 1)
		pks[0] = field
		return pks
	} else {
		//TODO: Return anonymous pk para len(pks) == 0
		return fieldsByTags("pk", str)
	}
}

func attributes(str reflect.Type, size int, pks []reflect.StructField) (a []reflect.StructField) {
	a = make([]reflect.StructField, size)
	count := 0

	for i := 0; i < str.NumField(); i++ {
		if str.Field(i).Name != "Id" && !matchField(pks, str.Field(i)) {
			a[count] = str.Field(i)
			count++
		}
	}

	return a
}

func fieldsByTags(tag string, str reflect.Type) (f []reflect.StructField) {
	f = make([]reflect.StructField, 0)

	for i := 0; i < str.NumField(); i++ {
		if strings.Contains(str.Field(i).Tag.Get("goe"), tag) {
			f = append(f, str.Field(i))
		}
	}
	return f
}

// matechField returns true if the field "t" is in the slice "s". otherwise false
func matchField(s []reflect.StructField, t reflect.StructField) bool {
	for i := range s {
		if reflect.DeepEqual(s[i], t) {
			return true
		}
	}
	return false
}

// checkMapping runs through the target fields and checks for nil fields
func checkMapping(target reflect.Value, str reflect.Type) {
	for i := 0; i < target.NumField(); i++ {
		if target.Field(i).IsNil() {
			log.Printf("goe: target field %q is nil on %q. try checking the struct %q for that field name",
				target.Type().Field(i).Name, target.Type(), str)
		}
	}
}
