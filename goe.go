package goe

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/lauro-ss/goe/utils"
)

func Open(db any, driver Driver) error {
	valueOf := reflect.ValueOf(db)
	if valueOf.Kind() != reflect.Ptr {
		return fmt.Errorf("%v: the target value needs to be pass as a pointer", pkg)
	}
	dbTarget := new(DB)
	valueOf = valueOf.Elem()

	dbTarget.addrMap = make(map[uintptr]field)

	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).IsNil() {
			valueOf.Field(i).Set(reflect.ValueOf(reflect.New(valueOf.Field(i).Type().Elem()).Interface()))
			initField(valueOf.Field(i).Elem(), dbTarget, driver)
		}
	}

	dbTarget.driver = driver
	dbTarget.driver.Init(dbTarget)
	valueOf.FieldByName("DB").Set(reflect.ValueOf(dbTarget))
	return nil
}

func initField(valueOf reflect.Value, db *DB, driver Driver) {
	p, fieldName := getPk(valueOf.Type(), driver)
	db.addrMap[uintptr(valueOf.FieldByName(fieldName).Addr().UnsafePointer())] = p
	var field reflect.StructField

	manyToOnes := make([]string, 0)
	for i := 0; i < valueOf.NumField(); i++ {
		field = valueOf.Type().Field(i)
		//skip primary key
		if field.Name == fieldName {
			continue
		}
		//skips many to one field
		if slices.Contains(manyToOnes, field.Name) {
			continue
		}
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			handlerSlice(valueOf.Field(i).Type().Elem(), valueOf, i, p, db, driver)
		case reflect.Struct:
			fk := handlerStruct(valueOf.Field(i).Type(), valueOf, i, p, db, driver)
			if fk != "" {
				manyToOnes = append(manyToOnes, fk)
				// remove fk field from attribute
				for k, v := range db.addrMap {
					if a, ok := v.(*att); ok && a.structAttributeName == fk {
						delete(db.addrMap, k)
					}
				}
			}
		case reflect.Ptr:
			if valueOf.Field(i).Type().Elem().Kind() == reflect.Struct {
				if mto := isManyToOne(valueOf.Field(i).Type().Elem(), valueOf.Type(), driver); mto != nil {
					key := driver.KeywordHandler(utils.TableNamePattern(valueOf.Field(i).Type().Elem().Name()))
					db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = mto
					mto.pk = p
					p.fks[key] = mto
					manyToOnes = append(manyToOnes, mto.idFkStructName)
					// remove fk field from attribute
					for k, v := range db.addrMap {
						if a, ok := v.(*att); ok && a.structAttributeName == mto.idFkStructName {
							delete(db.addrMap, k)
						}
					}
				}
			} else {
				newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
			}
		default:
			newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
		}
	}
}

func handlerStruct(targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *pk, db *DB, driver Driver) string {
	switch targetTypeOf.Name() {
	case "Time":
		newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
	default:
		if mto := isManyToOne(targetTypeOf, valueOf.Type(), driver); mto != nil {
			key := driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name()))
			db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = mto
			mto.pk = p
			p.fks[key] = mto
			return mto.idFkStructName
		}
	}
	return ""
}

func handlerSlice(targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *pk, db *DB, driver Driver) {
	switch targetTypeOf.Kind() {
	case reflect.Uint8:
		valueOf.Field(i).SetBytes([]byte{})
		newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
	default:
		if mtm := isManytoMany(targetTypeOf, valueOf.Type(), valueOf.Type().Field(i).Tag.Get("goe"), db, driver); mtm != nil {
			key := driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name()))
			p.fks[key] = mtm
		}
	}

}

func newAttr(valueOf reflect.Value, i int, p *pk, addr uintptr, db *DB, d Driver) {
	at := createAtt(
		valueOf.Type().Field(i).Name,
		p,
		d,
	)
	db.addrMap[addr] = at
}

func getPk(typeOf reflect.Type, driver Driver) (*pk, string) {
	var p *pk
	id, valid := typeOf.FieldByName("Id")
	if valid {
		p = createPk(typeOf.Name(), id.Name, isAutoIncrement(id), driver)
		return p, id.Name
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		//TODO: Set anonymous pk
		return nil, ""
	}
	p = createPk(typeOf.Name(), fields[0].Name, isAutoIncrement(fields[0]), driver)
	return p, fields[0].Name
}

func isAutoIncrement(id reflect.StructField) bool {
	return strings.Contains(id.Type.Kind().String(), "int")
}

func isManytoMany(targetTypeOf reflect.Type, typeOf reflect.Type, tag string, db *DB, driver Driver) any {
	nameTargetTypeOf := driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name()))
	nameTypeOf := driver.KeywordHandler(utils.TableNamePattern(typeOf.Name()))

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
				return createManyToMany(tag, typeOf, targetTypeOf, driver)
			}
		case reflect.Struct:
			if targetTypeOf.Field(i).Type.Name() == typeOf.Name() {
				return createManyToOne(typeOf, targetTypeOf, true, driver)
			}
		case reflect.Ptr:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				return createManyToOne(typeOf, targetTypeOf, true, driver)
			}
		}
	}

	return nil
}

func isManyToOne(targetTypeOf reflect.Type, typeOf reflect.Type, driver Driver) *manyToOne {
	for i := 0; i < targetTypeOf.NumField(); i++ {
		switch targetTypeOf.Field(i).Type.Kind() {
		case reflect.Slice:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				return createManyToOne(targetTypeOf, typeOf, false, driver)
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
			return after
		}
	}
	return ""
}
