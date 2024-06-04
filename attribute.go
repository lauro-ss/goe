package goe

import (
	"fmt"
	"reflect"
	"strings"
)

type manyToMany struct {
	table string
	ids   map[string]attributeStrings
}

func createManyToMany(tag string, typeOf reflect.Type, targetTypeOf reflect.Type) *manyToMany {
	table := getTagValue(tag, "table:")
	if table == "" {
		return nil
	}
	nameTargetTypeOf := strings.ToLower(targetTypeOf.Name())
	nameTypeOf := strings.ToLower(typeOf.Name())

	mtm := new(manyToMany)
	mtm.table = table
	mtm.ids = make(map[string]attributeStrings)
	pk := primaryKeys(typeOf)[0]

	id := pk.Name
	id += nameTypeOf
	mtm.ids[nameTypeOf] = createAttributeStrings(fmt.Sprintf("%v.%v", table, id), id, getType(pk))

	// target id
	pkTarget := primaryKeys(targetTypeOf)[0]
	id = pkTarget.Name
	id += nameTargetTypeOf

	mtm.ids[nameTargetTypeOf] = createAttributeStrings(fmt.Sprintf("%v.%v", table, id), id, getType(pkTarget))
	return mtm
}

type manyToOne struct {
	targetTable string
	nullable    bool
	id          string
	hasMany     bool
}

func createManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, hasMany bool, nullable bool) *manyToOne {
	mto := new(manyToOne)
	mto.targetTable = strings.ToLower(typeOf.Name())
	mto.id = fmt.Sprintf("%v.%v", strings.ToLower(targetTypeOf.Name()), strings.ToLower(primaryKeys(typeOf)[0].Name+typeOf.Name()))
	mto.hasMany = hasMany
	mto.nullable = nullable
	return mto
}

type attributeStrings struct {
	selectName    string
	attributeName string
	dataType      string
}

func createAttributeStrings(selectName string, attributeName string, dataType string) attributeStrings {
	return attributeStrings{
		selectName:    strings.ToLower(selectName),
		attributeName: strings.ToLower(attributeName),
		dataType:      strings.ToLower(dataType)}
}

type pk struct {
	table         string
	autoIncrement bool
	fks           map[string]any
	attributeStrings
}

func createPk(table string, selectName string, attributeName string, autoIncrement bool, typeString string) *pk {
	return &pk{
		table:            strings.ToLower(table),
		attributeStrings: createAttributeStrings(selectName, attributeName, typeString),
		autoIncrement:    autoIncrement,
		fks:              make(map[string]any)}
}

type att struct {
	attributeStrings
	nullable bool
	index    string
	pk       *pk
}

func createAtt(selectName string, attributeName string, pk *pk, typeString string, nullable bool, index string) *att {
	return &att{
		attributeStrings: createAttributeStrings(selectName, attributeName, typeString), pk: pk, nullable: nullable, index: index}
}
