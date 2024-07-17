package goe

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/lauro-ss/goe/utils"
)

var ErrInvalidManyToOne = errors.New("goe")

func Open(db any, driver Driver) error {
	valueOf := reflect.ValueOf(db)
	if valueOf.Kind() != reflect.Ptr {
		return fmt.Errorf("%v: the target value needs to be pass as a pointer", pkg)
	}
	dbTarget := new(DB)
	valueOf = valueOf.Elem()

	dbTarget.addrMap = make(map[uintptr]field)

	// set value for fields
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).IsNil() {
			valueOf.Field(i).Set(reflect.ValueOf(reflect.New(valueOf.Field(i).Type().Elem()).Interface()))
		}
	}

	var err error
	// init fields
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).Elem().Type().Name() != "DB" {
			err = initField(valueOf, valueOf.Field(i).Elem(), dbTarget, driver)
			if err != nil {
				return err
			}
		}
	}

	dbTarget.driver = driver
	dbTarget.driver.Init(dbTarget)
	valueOf.FieldByName("DB").Set(reflect.ValueOf(dbTarget))
	return nil
}

func initField(tables reflect.Value, valueOf reflect.Value, db *DB, driver Driver) error {
	p, fieldName := getPk(valueOf.Type(), driver)
	db.addrMap[uintptr(valueOf.FieldByName(fieldName).Addr().UnsafePointer())] = p
	var field reflect.StructField

	for i := 0; i < valueOf.NumField(); i++ {
		field = valueOf.Type().Field(i)
		//skip primary key
		if field.Name == fieldName {
			continue
		}
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			err := handlerSlice(tables, valueOf.Field(i).Type().Elem(), valueOf, i, p, db, driver)
			if err != nil {
				return err
			}
		case reflect.Struct:
			handlerStruct(valueOf.Field(i).Type(), valueOf, i, p, db, driver)
		case reflect.Ptr:
			table := getTagValue(valueOf.Type().Field(i).Tag.Get("goe"), "table:")
			if table != "" {
				if mto := isManyToOne(tables, valueOf.Type(), table, driver); mto != nil {
					switch v := mto.(type) {
					case *manyToOne:
						key := driver.KeywordHandler(utils.TableNamePattern(table))
						db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
						v.pk = p
						p.fks[key] = v
					case *oneToOne:
						key := driver.KeywordHandler(utils.TableNamePattern(table))
						db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
						v.pk = p
						p.fks[key] = v
					}

				} else {
					return fmt.Errorf("%w: field %q on %q has table %q specified but the table don't exists",
						ErrInvalidManyToOne,
						valueOf.Type().Field(i).Name,
						valueOf.Type().Name(),
						table)
				}
			} else {
				newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
			}
		default:
			// check if tag field exists for many to one
			table := getTagValue(valueOf.Type().Field(i).Tag.Get("goe"), "table:")
			if table != "" {
				if mto := isManyToOne(tables, valueOf.Type(), table, driver); mto != nil {
					switch v := mto.(type) {
					case *manyToOne:
						key := driver.KeywordHandler(utils.TableNamePattern(table))
						db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
						v.pk = p
						p.fks[key] = v
					case *oneToOne:
						key := driver.KeywordHandler(utils.TableNamePattern(table))
						db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
						v.pk = p
						p.fks[key] = v
					}
				} else {
					return fmt.Errorf("%w: field %q on %q has table %q specified but the table don't exists",
						ErrInvalidManyToOne,
						valueOf.Type().Field(i).Name,
						valueOf.Type().Name(),
						table)
				}
			} else {
				newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
			}
		}
	}
	return nil
}

func handlerStruct(targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *pk, db *DB, driver Driver) {
	switch targetTypeOf.Name() {
	case "Time":
		newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
	}
}

func handlerSlice(tables reflect.Value, targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *pk, db *DB, driver Driver) error {
	switch targetTypeOf.Kind() {
	case reflect.Uint8:
		table := getTagValue(valueOf.Type().Field(i).Tag.Get("goe"), "table:")
		if table != "" {
			if mto := isManyToOne(tables, valueOf.Type(), table, driver); mto != nil {
				switch v := mto.(type) {
				case *manyToOne:
					key := driver.KeywordHandler(utils.TableNamePattern(table))
					db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
					v.pk = p
					p.fks[key] = v
				case *oneToOne:
					key := driver.KeywordHandler(utils.TableNamePattern(table))
					db.addrMap[uintptr(valueOf.Field(i).Addr().UnsafePointer())] = v
					v.pk = p
					p.fks[key] = v
				}
			} else {
				return fmt.Errorf("%w: field %q on %q has table %q specified but the table don't exists",
					ErrInvalidManyToOne,
					valueOf.Type().Field(i).Name,
					valueOf.Type().Name(),
					table)
			}
		}
		valueOf.Field(i).SetBytes([]byte{})
		newAttr(valueOf, i, p, uintptr(valueOf.Field(i).Addr().UnsafePointer()), db, driver)
	default:
		if mtm := isManytoMany(targetTypeOf, valueOf.Type(), valueOf.Type().Field(i).Tag.Get("goe"), db, driver); mtm != nil {
			key := driver.KeywordHandler(utils.TableNamePattern(targetTypeOf.Name()))
			p.fks[key] = mtm
		}
	}
	return nil
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
		default:
			if typeOf.Name() == getTagValue(targetTypeOf.Field(i).Tag.Get("goe"), "table:") {
				return createManyToOne(typeOf, targetTypeOf, true, driver)
			}
		}
	}

	return nil
}

func isManyToOne(tables reflect.Value, typeOf reflect.Type, table string, driver Driver) field {
	for c := 0; c < tables.NumField(); c++ {
		if tables.Field(c).Elem().Type().Name() == table {
			for i := 0; i < tables.Field(c).Elem().NumField(); i++ {
				// check if there is a slice to typeOf
				if tables.Field(c).Elem().Field(i).Kind() == reflect.Slice {
					if tables.Field(c).Elem().Field(i).Type().Elem().Name() == typeOf.Name() {
						return createManyToOne(tables.Field(c).Elem().Type(), typeOf, false, driver)
					}
				}
			}
			return createOneToOne(tables.Field(c).Elem().Type(), typeOf, driver)
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
