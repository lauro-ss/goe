package goe

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func Open(db any, driverName string, uri string) error {
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

	dbTarget.open(driverName, uri)
	valueOf.FieldByName("DB").Set(reflect.ValueOf(dbTarget))
	return nil
}

func initField(valueOf reflect.Value, db *DB) {
	p, fieldName := getPk(valueOf.Type())

	db.addrMap[fmt.Sprint(valueOf.FieldByName(fieldName).Addr())] = p

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
			if valueOf.Type().Field(i).Name != fieldName {
				at := createAtt(
					fmt.Sprintf("%v.%v", valueOf.Type().Name(), valueOf.Type().Field(i).Name),
					valueOf.Type().Field(i).Name,
					p)
				db.addrMap[fmt.Sprint(valueOf.Field(i).Addr())] = at
			}
		}
	}
}

func getPk(typeOf reflect.Type) (*pk, string) {
	var p *pk
	id, valid := typeOf.FieldByName("Id")
	if valid {
		p = createPk(typeOf.Name(), typeOf.Name()+"."+id.Name, id.Name)
		return p, id.Name
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		//Set anonymous pk
		return nil, ""
	}
	p = createPk(typeOf.Name(), typeOf.Name()+"."+fields[0].Name, fields[0].Name)
	return p, fields[0].Name
}

func isManytoMany(targetTypeOf reflect.Type, typeOf reflect.Type, tag string, db *DB) any {
	for _, v := range db.addrMap {
		switch value := v.(type) {
		case *pk:
			if value.table == targetTypeOf.Name() {
				switch fk := value.fks[targetTypeOf.Name()].(type) {
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
				re := regexp.MustCompile("table:")
				result := re.Split(tag, 2)
				if len(result) == 0 {
					return nil
				}
				table := re.Split(tag, 2)[1]
				var mtm manyToMany
				mtm.table = table
				mtm.ids = make(map[string]string)

				id := primaryKeys(typeOf)[0].Name
				id += typeOf.Name()
				mtm.ids[typeOf.Name()] = fmt.Sprintf("%v.%v", table, id)

				// target id
				id = primaryKeys(targetTypeOf)[0].Name
				id += targetTypeOf.Name()

				mtm.ids[targetTypeOf.Name()] = fmt.Sprintf("%v.%v", table, id)

				return &mtm
			}
		case reflect.Struct:
			if targetTypeOf.Field(i).Type.Name() == typeOf.Name() {
				var mto manyToOne
				mto.id = fmt.Sprintf("%v.%v", targetTypeOf.Name(), primaryKeys(typeOf)[0].Name+typeOf.Name())
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
				mto.id = fmt.Sprintf("%v.%v", typeOf.Name(), primaryKeys(targetTypeOf)[0].Name+targetTypeOf.Name())
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
