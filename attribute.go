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
	mtm.ids[nameTypeOf] = createAttributeStrings(fmt.Sprintf("%v.%v", table, id), id)

	// target id
	pkTarget := primaryKeys(targetTypeOf)[0]
	id = pkTarget.Name
	id += nameTargetTypeOf

	mtm.ids[nameTargetTypeOf] = createAttributeStrings(fmt.Sprintf("%v.%v", table, id), id)
	return mtm
}

type manyToOne struct {
	targetTable string
	id          string
	hasMany     bool
}

func createManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, hasMany bool) *manyToOne {
	mto := new(manyToOne)
	mto.targetTable = strings.ToLower(typeOf.Name())
	mto.id = fmt.Sprintf("%v.%v", strings.ToLower(targetTypeOf.Name()), strings.ToLower(primaryKeys(typeOf)[0].Name+typeOf.Name()))
	mto.hasMany = hasMany
	return mto
}

type attributeStrings struct {
	selectName    string
	attributeName string
}

func createAttributeStrings(selectName string, attributeName string) attributeStrings {
	return attributeStrings{
		selectName:    strings.ToLower(selectName),
		attributeName: strings.ToLower(attributeName),
	}
}

type pk struct {
	table         string
	autoIncrement bool
	fks           map[string]any
	attributeStrings
}

func createPk(table string, selectName string, attributeName string, autoIncrement bool) *pk {
	return &pk{
		table:            strings.ToLower(table),
		attributeStrings: createAttributeStrings(selectName, attributeName),
		autoIncrement:    autoIncrement,
		fks:              make(map[string]any)}
}

type att struct {
	attributeStrings
	pk *pk
}

func createAtt(selectName string, attributeName string, pk *pk) *att {
	return &att{
		attributeStrings: createAttributeStrings(selectName, attributeName), pk: pk}
}
