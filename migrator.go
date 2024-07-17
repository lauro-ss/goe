package goe

import (
	"fmt"
	"reflect"

	"github.com/lauro-ss/goe/utils"
)

type Migrator struct {
	Tables []any
}

func MigrateFrom(db any) *Migrator {
	valueOf := reflect.ValueOf(db).Elem()

	migrator := new(Migrator)
	migrator.Tables = make([]any, 0)
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).Type().Elem().Name() != "DB" {
			typeField(valueOf, valueOf.Field(i).Elem(), migrator)
		}
	}

	return migrator
}

func typeField(tables reflect.Value, valueOf reflect.Value, migrator *Migrator) {
	p, fieldName := migratePk(valueOf.Type())

	migrator.Tables = append(migrator.Tables, p)
	var field reflect.StructField

	for i := 0; i < valueOf.NumField(); i++ {
		field = valueOf.Type().Field(i)
		//skip primary key
		if field.Name == fieldName {
			continue
		}
		switch valueOf.Field(i).Kind() {
		case reflect.Slice:
			handlerSliceMigrate(tables, field, valueOf.Field(i).Type().Elem(), valueOf, i, p, migrator)
		case reflect.Struct:
			handlerStructMigrate(field, valueOf.Field(i).Type(), valueOf, i, p, migrator)
		case reflect.Ptr:
			table := getTagValue(valueOf.Type().Field(i).Tag.Get("goe"), "table:")
			if table != "" {
				if mto := isMigrateManyToOne(tables, table, valueOf.Type(), true); mto != nil {
					key := utils.TableNamePattern(table)
					p.Fks[key] = mto
				}
				//TODO: Add fk error
				continue
			}
			migrateAtt(valueOf, field, i, p, migrator)
		default:
			table := getTagValue(valueOf.Type().Field(i).Tag.Get("goe"), "table:")
			if table != "" {
				if mto := isMigrateManyToOne(tables, table, valueOf.Type(), false); mto != nil {
					key := utils.TableNamePattern(table)
					p.Fks[key] = mto
				}
				//TODO: Add fk error
				continue
			}
			migrateAtt(valueOf, field, i, p, migrator)
		}
	}
}

func handlerStructMigrate(field reflect.StructField, targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *MigratePk, migrator *Migrator) {
	switch targetTypeOf.Name() {
	case "Time":
		migrateAtt(valueOf, field, i, p, migrator)
	}
}

func handlerSliceMigrate(tables reflect.Value, field reflect.StructField, targetTypeOf reflect.Type, valueOf reflect.Value, i int, p *MigratePk, migrator *Migrator) {
	switch targetTypeOf.Kind() {
	case reflect.Uint8:
		table := getTagValue(valueOf.Type().Field(i).Tag.Get("goe"), "table:")
		if table != "" {
			if mto := isMigrateManyToOne(tables, table, valueOf.Type(), false); mto != nil {
				key := utils.TableNamePattern(table)
				p.Fks[key] = mto
			}
			//TODO: Add fk error
			return
		}
		migrateAtt(valueOf, field, i, p, migrator)
	default:
		if mtm := isMigrateManytoMany(targetTypeOf, valueOf.Type(), valueOf.Type().Field(i).Tag.Get("goe"), migrator); mtm != nil {
			key := utils.TableNamePattern(targetTypeOf.Name())
			p.Fks[key] = mtm
		}
	}

}

func isMigrateManyToOne(tables reflect.Value, table string, typeOf reflect.Type, nullable bool) any {
	for c := 0; c < tables.NumField(); c++ {
		if tables.Field(c).Elem().Type().Name() == table {
			for i := 0; i < tables.Field(c).Elem().NumField(); i++ {
				// check if there is a slice to typeOf
				if tables.Field(c).Elem().Field(i).Kind() == reflect.Slice {
					if tables.Field(c).Elem().Field(i).Type().Elem().Name() == typeOf.Name() {
						return createMigrateManyToOne(tables.Field(c).Elem().Type(), typeOf, false, nullable)
					}
				}
			}
			return createMigrateOneToOne(tables.Field(c).Elem().Type(), typeOf, nullable)
		}
	}
	return nil
}

func isMigrateManytoMany(targetTypeOf reflect.Type, typeOf reflect.Type, tag string, m *Migrator) any {
	nameTargetTypeOf := utils.TableNamePattern(targetTypeOf.Name())
	nameTypeOf := utils.TableNamePattern(typeOf.Name())

	for _, v := range m.Tables {
		switch value := v.(type) {
		case *MigratePk:
			if value.Table == nameTargetTypeOf {
				switch fk := value.Fks[nameTypeOf].(type) {
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
	nameTargetTypeOf := targetTypeOf.Name()
	nameTypeOf := typeOf.Name()

	mtm := new(MigrateManyToMany)
	mtm.Table = utils.TableNamePattern(table)
	mtm.Ids = make(map[string]AttributeStrings)
	pk := primaryKeys(typeOf)[0]

	id := utils.ManyToManyNamePattern(pk.Name, nameTypeOf)
	mtm.Ids[utils.TableNamePattern(nameTypeOf)] = setAttributeStrings(id, getType(pk))

	// target id
	pkTarget := primaryKeys(targetTypeOf)[0]
	id = utils.ManyToManyNamePattern(pkTarget.Name, nameTargetTypeOf)

	mtm.Ids[utils.TableNamePattern(nameTargetTypeOf)] = setAttributeStrings(id, getType(pkTarget))
	return mtm
}

func createMigrateManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, hasMany bool, nullable bool) *MigrateManyToOne {
	mto := new(MigrateManyToOne)
	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.Id = fmt.Sprintf("%v.%v", utils.TableNamePattern(targetTypeOf.Name()), utils.ManyToOneNamePattern(primaryKeys(typeOf)[0].Name, typeOf.Name()))
	mto.HasMany = hasMany
	mto.Nullable = nullable
	return mto
}

func createMigrateOneToOne(typeOf reflect.Type, targetTypeOf reflect.Type, nullable bool) *MigrateOneToOne {
	mto := new(MigrateOneToOne)
	mto.TargetTable = utils.TableNamePattern(typeOf.Name())
	mto.Id = fmt.Sprintf("%v.%v", utils.TableNamePattern(targetTypeOf.Name()), utils.ManyToOneNamePattern(primaryKeys(typeOf)[0].Name, typeOf.Name()))
	mto.Nullable = nullable
	return mto
}

type MigratePk struct {
	Table         string
	AutoIncrement bool
	Fks           map[string]any
	AttributeName string
	DataType      string
}

type MigrateAtt struct {
	Nullable      bool
	Index         string
	Pk            *MigratePk
	AttributeName string
	DataType      string
}

type MigrateOneToOne struct {
	TargetTable string
	Nullable    bool
	Id          string
}

type MigrateManyToOne struct {
	TargetTable string
	Nullable    bool
	Id          string
	HasMany     bool
}

type MigrateManyToMany struct {
	Table string
	Ids   map[string]AttributeStrings
}

type AttributeStrings struct {
	AttributeName string
	DataType      string
}

func setAttributeStrings(attributeName string, dataType string) AttributeStrings {
	return AttributeStrings{
		AttributeName: attributeName,
		DataType:      dataType}
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
		Table:         utils.TableNamePattern(table),
		AttributeName: utils.ColumnNamePattern(attributeName),
		DataType:      dataType,
		AutoIncrement: autoIncrement,
		Fks:           make(map[string]any)}
}

func createMigrateAtt(attributeName string, pk *MigratePk, dataType string, nullable bool, index string) *MigrateAtt {
	return &MigrateAtt{
		AttributeName: utils.ColumnNamePattern(attributeName),
		DataType:      dataType,
		Pk:            pk,
		Nullable:      nullable,
		Index:         index,
	}
}
