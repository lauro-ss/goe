package goe

import (
	"fmt"
	"reflect"
	"strings"
)

type Migrator struct {
	Tables []any
}

func Migrate(db any) *Migrator {
	valueOf := reflect.ValueOf(db).Elem()

	migrator := new(Migrator)
	migrator.Tables = make([]any, 0)
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).Type().Elem().Name() != "DB" {
			typeField(valueOf.Field(i).Elem(), migrator)
		}
	}

	return migrator
}

func typeField(valueOf reflect.Value, migrator *Migrator) {
	p, fieldName := migratePk(valueOf.Type())

	migrator.Tables = append(migrator.Tables, p)
	var field reflect.StructField

	for i := 0; i < valueOf.NumField(); i++ {
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			if mtm := isMigrateManytoMany(valueOf.Field(i).Type().Elem(), valueOf.Type(), valueOf.Type().Field(i).Tag.Get("goe"), migrator); mtm != nil {
				key := strings.ToLower(valueOf.Field(i).Type().Elem().Name()) //TODO: add to lower only method
				p.fks[key] = mtm
			}
		case reflect.Struct:
			if mto := isMigrateManyToOne(valueOf.Field(i).Type(), valueOf.Type(), false); mto != nil {
				key := strings.ToLower(valueOf.Field(i).Type().Name()) //TODO: add to lower only method
				p.fks[key] = mto
			}
		case reflect.Ptr:
			if valueOf.Field(i).Type().Elem().Kind() == reflect.Struct {
				if mto := isMigrateManyToOne(valueOf.Field(i).Type().Elem(), valueOf.Type(), true); mto != nil {
					key := strings.ToLower(valueOf.Field(i).Type().Elem().Name())
					p.fks[key] = mto
				}
			} else {
				field = valueOf.Type().Field(i)
				if field.Name != fieldName {
					migrateAtt(valueOf, field, i, p, migrator)
				}
			}
		default:
			field = valueOf.Type().Field(i)
			if field.Name != fieldName {
				migrateAtt(valueOf, field, i, p, migrator)
			}
		}
	}
}

func isMigrateManyToOne(targetTypeOf reflect.Type, typeOf reflect.Type, nullable bool) *MigrateManyToOne {
	for i := 0; i < targetTypeOf.NumField(); i++ {
		switch targetTypeOf.Field(i).Type.Kind() {
		case reflect.Slice:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				return createMigrateManyToOne(targetTypeOf, typeOf, false, nullable)
			}
		}
	}

	return nil
}

func isMigrateManytoMany(targetTypeOf reflect.Type, typeOf reflect.Type, tag string, m *Migrator) any {
	nameTargetTypeOf := strings.ToLower(targetTypeOf.Name())
	nameTypeOf := strings.ToLower(typeOf.Name())

	for _, v := range m.Tables {
		switch value := v.(type) {
		case *MigratePk:
			if value.table == nameTargetTypeOf {
				switch fk := value.fks[nameTypeOf].(type) {
				case *MigrateManyToMany:
					return fk
				}
			}
		}
	}

	for i := 0; i < targetTypeOf.NumField(); i++ {
		switch targetTypeOf.Field(i).Type.Kind() {
		case reflect.Slice:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				return createMigrateManyToMany(tag, typeOf, targetTypeOf)
			}
		case reflect.Struct:
			if targetTypeOf.Field(i).Type.Name() == typeOf.Name() {
				return createMigrateManyToOne(typeOf, targetTypeOf, true, false)
			}
		case reflect.Ptr:
			if targetTypeOf.Field(i).Type.Elem().Name() == typeOf.Name() {
				return createMigrateManyToOne(typeOf, targetTypeOf, true, false)
			}
		}
	}

	return nil
}

func createMigrateManyToMany(tag string, typeOf reflect.Type, targetTypeOf reflect.Type) *MigrateManyToMany {
	table := getTagValue(tag, "table:")
	if table == "" {
		return nil
	}
	nameTargetTypeOf := strings.ToLower(targetTypeOf.Name())
	nameTypeOf := strings.ToLower(typeOf.Name())

	mtm := new(MigrateManyToMany)
	mtm.table = table
	mtm.ids = make(map[string]AttributeStrings)
	pk := primaryKeys(typeOf)[0]

	id := pk.Name
	id += nameTypeOf
	mtm.ids[nameTypeOf] = setAttributeStrings(id, getType(pk))

	// target id
	pkTarget := primaryKeys(targetTypeOf)[0]
	id = pkTarget.Name
	id += nameTargetTypeOf

	mtm.ids[nameTargetTypeOf] = setAttributeStrings(id, getType(pkTarget))
	return mtm
}

func createMigrateManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, hasMany bool, nullable bool) *MigrateManyToOne {
	mto := new(MigrateManyToOne)
	mto.targetTable = strings.ToLower(typeOf.Name())
	mto.id = fmt.Sprintf("%v.%v", strings.ToLower(targetTypeOf.Name()), strings.ToLower(primaryKeys(typeOf)[0].Name+typeOf.Name()))
	mto.hasMany = hasMany
	mto.nullable = nullable
	return mto
}

type MigratePk struct {
	table         string
	autoIncrement bool
	fks           map[string]any
	attributeName string
	dataType      string
}

type MigrateAtt struct {
	nullable      bool
	index         string
	pk            *MigratePk
	attributeName string
	dataType      string
}

type MigrateManyToOne struct {
	targetTable string
	nullable    bool
	id          string
	hasMany     bool
}

type MigrateManyToMany struct {
	table string
	ids   map[string]AttributeStrings
}

type AttributeStrings struct {
	attributeName string
	dataType      string
}

func setAttributeStrings(attributeName string, dataType string) AttributeStrings {
	return AttributeStrings{
		attributeName: strings.ToLower(attributeName),
		dataType:      strings.ToLower(dataType)}
}

func migratePk(typeOf reflect.Type) (*MigratePk, string) {
	var p *MigratePk
	id, valid := typeOf.FieldByName("Id")
	if valid {
		p = createMigratePk(typeOf.Name(), id.Name, isAutoIncrement(id), getType(id))
		return p, id.Name
	}

	fields := fieldsByTags("pk", typeOf)
	if len(fields) == 0 {
		//Set anonymous pk
		return nil, ""
	}
	p = createMigratePk(typeOf.Name(), fields[0].Name, isAutoIncrement(fields[0]), getType(fields[0]))
	return p, fields[0].Name
}

func migrateAtt(valueOf reflect.Value, field reflect.StructField, i int, pk *MigratePk, m *Migrator) {
	at := createMigrateAtt(
		valueOf.Type().Field(i).Name,
		pk,
		getType(field),
		field.Type.String()[0] == '*',
		getIndex(field),
	)
	m.Tables = append(m.Tables, at)
	//db.addrMap[fmt.Sprint(valueOf.Field(i).Addr())] = at TODO::
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

func getIndex(field reflect.StructField) string {
	value := getTagValue(field.Tag.Get("goe"), "index(")
	if value != "" {
		return value[0 : len(value)-1]
	}
	return ""
}

func createMigratePk(table string, attributeName string, autoIncrement bool, dataType string) *MigratePk {
	return &MigratePk{
		table:         strings.ToLower(table),
		attributeName: strings.ToLower(attributeName),
		dataType:      dataType,
		autoIncrement: autoIncrement,
		fks:           make(map[string]any)}
}

func createMigrateAtt(attributeName string, pk *MigratePk, dataType string, nullable bool, index string) *MigrateAtt {
	return &MigrateAtt{
		attributeName: strings.ToLower(attributeName),
		dataType:      dataType,
		pk:            pk,
		nullable:      nullable,
		index:         index,
	}
}
