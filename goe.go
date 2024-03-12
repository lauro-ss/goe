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
	mapData(database, target, str)
	checkMapping(target, str)
	return nil
}

// Map makes the mapping from the struct fields "s" to the target "t"
func mapData(database reflect.Value, target reflect.Value, str reflect.Type) {
	pks := primaryKeys(str)

	var pk *Pk
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
				mapForeignKey(database, pk, attr)
			}
		}
	}
}

func mapPrimaryKey(target reflect.Value, field reflect.StructField, tableName string) *Pk {
	if target.Kind() == reflect.Invalid {
		return nil
	}

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

func mapForeignKey(database reflect.Value, pk *Pk, field reflect.StructField) {
	switch field.Type.Kind() {
	case reflect.Struct:
		//many to one
		fmt.Println("Struct")
	case reflect.Slice:
		//possibile many to many
		str := field.Type.Elem()

		//TODO: Add more then one primary key
		target := database.FieldByName(str.Name()).FieldByName(primaryKeys(str)[0].Name)
		if target.Kind() == reflect.Invalid {
			return
		}

		if target.IsZero() {
			field := primaryKeys(str)[0]
			pk.Fk[str.Name()] = mapPrimaryKey(target, field, str.Name())
		} else {
			pk.Fk[str.Name()] = getPrimaryKey(database, str)
		}
	}

}

func getPrimaryKey(database reflect.Value, str reflect.Type) *Pk {
	//TODO: Add more then one primary key
	field := database.FieldByName(str.Name()).FieldByName(primaryKeys(str)[0].Name)
	return (*Pk)(field.UnsafePointer())
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
