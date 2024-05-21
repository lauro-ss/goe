package goe

import "strings"

type manyToMany struct {
	table string
	ids   map[string]attributeStrings
}

type manyToOne struct {
	id      string
	hasMany bool
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
	pk       *pk
}

func createAtt(selectName string, attributeName string, pk *pk, typeString string, nullable bool) *att {
	return &att{
		attributeStrings: createAttributeStrings(selectName, attributeName, typeString), pk: pk, nullable: nullable}
}
