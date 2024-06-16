package goe

import (
	"fmt"
	"reflect"
	"strings"
)

func Open(db any, driver Driver) error {
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

	dbTarget.driver = driver
	dbTarget.driver.Init(dbTarget)
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
				key := strings.ToLower(valueOf.Field(i).Type().Elem().Name())
				p.fks[key] = mtm
			}
		case reflect.Struct:
			if mto := isManyToOne(valueOf.Field(i).Type(), valueOf.Type()); mto != nil {
				key := strings.ToLower(valueOf.Field(i).Type().Name())
				p.fks[key] = mto
			}
		case reflect.Ptr:
			if valueOf.Field(i).Type().Elem().Kind() == reflect.Struct {
				if mto := isManyToOne(valueOf.Field(i).Type().Elem(), valueOf.Type()); mto != nil {
					key := strings.ToLower(valueOf.Field(i).Type().Elem().Name())
					p.fks[key] = mto
				}
			} else {
				field = valueOf.Type().Field(i)
				if field.Name != fieldName {
					newAttr(valueOf, i, p, db)
				}
			}
		default:
			field = valueOf.Type().Field(i)
			if field.Name != fieldName {
				newAttr(valueOf, i, p, db)
			}
		}
	}
}

func newAttr(valueOf reflect.Value, i int, p *pk, db *DB) {
	at := createAtt(
		fmt.Sprintf("%v.%v", valueOf.Type().Name(), valueOf.Type().Field(i).Name),
		valueOf.Type().Field(i).Name,
		p,
	)
	db.addrMap[fmt.Sprint(valueOf.Field(i).Addr())] = at
}

func getPk(typeOf reflect.Type) (*pk, string) {
	var p *pk
	id, valid := typeOf.FieldByName("Id")
	if valid {
		p = createPk(typeOf.Name(), typeOf.Name()+"."+id.Name, id.Name, isAutoIncrement(id))
		return p, id.Name
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		//Set anonymous pk
		return nil, ""
	}
	p = createPk(typeOf.Name(), typeOf.Name()+"."+fields[0].Name, fields[0].Name, isAutoIncrement(fields[0]))
	return p, fields[0].Name
}

func isAutoIncrement(id reflect.StructField) bool {
	return strings.Contains(id.Type.Kind().String(), "int")
}

func isManytoMany(targetTypeOf reflect.Type, typeOf reflect.Type, tag string, db *DB) any {
	nameTargetTypeOf := strings.ToLower(targetTypeOf.Name())
	nameTypeOf := strings.ToLower(typeOf.Name())

	for _, v := range db.addrMap {
		switch value := v.(type) {
		case *pk:
			if value.table == nameTargetTypeOf {
				switch fk := value.fks[nameTypeOf].(type) {
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
				return createManyToMany(tag, typeOf, targetTypeOf)
			}
		case reflect.Struct:
			if targetTypeOf.Field(i).Type.Name() == typeOf.Name() {
				return createManyToOne(typeOf, targetTypeOf, true)
			}
		case reflect.Ptr:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				return createManyToOne(typeOf, targetTypeOf, true)
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
				return createManyToOne(targetTypeOf, typeOf, false)
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
		if after, found := strings.CutPrefix(v, subTag); found {
			return strings.ToLower(after)
		}
	}
	return ""
}
