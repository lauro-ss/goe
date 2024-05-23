package goe

import (
	"fmt"
	"reflect"
	"strings"
)

type Config struct {
	MigrationsPath string
}

func Open(db any, driverName string, uri string, config Config) error {
	valueOf := reflect.ValueOf(db)
	if valueOf.Kind() != reflect.Ptr {
		return fmt.Errorf("%v: the target value needs to be pass as a pointer", pkg)
	}
	var dbTarget *DB
	valueOf = valueOf.Elem()

	for i := 0; i < valueOf.NumField(); i++ {
		if !valueOf.Field(i).IsNil() {
			dbTarget = (reflect.New(valueOf.Field(i).Type().Elem()).Interface()).(*DB)
		}
	}
	dbTarget.addrMap = make(map[string]any)

	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).IsNil() {
			valueOf.Field(i).Set(reflect.ValueOf(reflect.New(valueOf.Field(i).Type().Elem()).Interface()))
			initField(valueOf.Field(i).Elem(), dbTarget)
		}
	}

	err := dbTarget.open(driverName, uri)
	if err != nil {
		return err
	}
	dbTarget.config = config
	valueOf.FieldByName("DB").Set(reflect.ValueOf(dbTarget))
	return nil
}

func initField(valueOf reflect.Value, db *DB) {
	p, fieldName := getPk(valueOf.Type())

	db.addrMap[fmt.Sprint(valueOf.FieldByName(fieldName).Addr())] = p
	var field reflect.StructField

	for i := 0; i < valueOf.NumField(); i++ {
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			if mtm := isManytoMany(valueOf.Field(i).Type().Elem(), valueOf.Type(), valueOf.Type().Field(i).Tag.Get("goe"), db); mtm != nil {
				p.fks[valueOf.Field(i).Type().Elem().Name()] = mtm
			}
		case reflect.Struct:
			if mto := isManyToOne(valueOf.Field(i).Type(), valueOf.Type()); mto != nil {
				p.fks[valueOf.Field(i).Type().Name()] = mto
			}
		default:
			field = valueOf.Type().Field(i)
			if field.Name != fieldName {
				at := createAtt(
					fmt.Sprintf("%v.%v", valueOf.Type().Name(), valueOf.Type().Field(i).Name),
					valueOf.Type().Field(i).Name,
					p,
					getType(field),
					field.Type.String()[0] == '*',
					getIndex(field),
				)
				db.addrMap[fmt.Sprint(valueOf.Field(i).Addr())] = at
			}
		}
	}
}

func getPk(typeOf reflect.Type) (*pk, string) {
	var p *pk
	id, valid := typeOf.FieldByName("Id")
	if valid {
		p = createPk(typeOf.Name(), typeOf.Name()+"."+id.Name, id.Name, isAutoIncrement(id), getType(id))
		return p, id.Name
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		//Set anonymous pk
		return nil, ""
	}
	p = createPk(typeOf.Name(), typeOf.Name()+"."+fields[0].Name, fields[0].Name, isAutoIncrement(fields[0]), getType(fields[0]))
	return p, fields[0].Name
}

func isAutoIncrement(id reflect.StructField) bool {
	return id.Type.Kind() == reflect.Int
}

func isManytoMany(targetTypeOf reflect.Type, typeOf reflect.Type, tag string, db *DB) any {
	for _, v := range db.addrMap {
		switch value := v.(type) {
		case *pk:
			if value.table == targetTypeOf.Name() {
				switch fk := value.fks[typeOf.Name()].(type) {
				case *manyToMany:
					return fk
				}
			}
		}
	}

	for i := 0; i < targetTypeOf.NumField(); i++ {
		switch targetTypeOf.Field(i).Type.Kind() {
		case reflect.Slice:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				table := getTagValue(tag, "table:")
				if table == "" {
					return nil
				}
				var mtm manyToMany
				mtm.table = table
				mtm.ids = make(map[string]attributeStrings)

				pk := primaryKeys(typeOf)[0]
				id := pk.Name
				id += typeOf.Name()
				mtm.ids[typeOf.Name()] = createAttributeStrings(fmt.Sprintf("%v.%v", table, id), id, getType(pk))

				// target id
				pkTarget := primaryKeys(targetTypeOf)[0]
				id = pkTarget.Name
				id += targetTypeOf.Name()

				mtm.ids[targetTypeOf.Name()] = createAttributeStrings(fmt.Sprintf("%v.%v", table, id), id, getType(pkTarget))

				return &mtm
			}
		case reflect.Struct:
			if targetTypeOf.Field(i).Type.Name() == typeOf.Name() {
				var mto manyToOne
				mto.id = fmt.Sprintf("%v.%v", strings.ToLower(targetTypeOf.Name()), strings.ToLower(primaryKeys(typeOf)[0].Name+typeOf.Name()))
				mto.hasMany = true
				return &mto
			}
		}
	}

	return nil
}

func isManyToOne(targetTypeOf reflect.Type, typeOf reflect.Type) *manyToOne {
	for i := 0; i < targetTypeOf.NumField(); i++ {
		switch targetTypeOf.Field(i).Type.Kind() {
		case reflect.Slice:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				var mto manyToOne
				mto.id = fmt.Sprintf("%v.%v", strings.ToLower(typeOf.Name()), strings.ToLower(primaryKeys(targetTypeOf)[0].Name+targetTypeOf.Name()))
				mto.hasMany = false

				return &mto
			}
		}
	}

	return nil
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

func fieldsByTags(tag string, str reflect.Type) (f []reflect.StructField) {
	f = make([]reflect.StructField, 0)

	for i := 0; i < str.NumField(); i++ {
		if strings.Contains(str.Field(i).Tag.Get("goe"), tag) {
			f = append(f, str.Field(i))
		}
	}
	return f
}

func getTagValue(fieldTag string, subTag string) string {
	values := strings.Split(fieldTag, ";")
	for _, v := range values {
		if _, after, found := strings.Cut(v, subTag); found {
			return after
		}
	}
	return ""
}

func getType(field reflect.StructField) string {
	value := getTagValue(field.Tag.Get("goe"), "type:")
	if value != "" {
		return value
	}
	dataType := field.Type.String()
	if dataType[0] == '*' {
		return dataType[1:]
	}
	return dataType
}

func getIndex(field reflect.StructField) bool {
	value := getTagValue(field.Tag.Get("goe"), "index:")
	return value == "unique"
}
