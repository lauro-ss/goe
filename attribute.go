package goe

import (
	"fmt"
	"reflect"

	"github.com/lauro-ss/goe/utils"
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
	nameTargetTypeOf := targetTypeOf.Name()
	nameTypeOf := typeOf.Name()

	mtm := new(manyToMany)
	mtm.table = utils.TableNamePattern(table)
	mtm.ids = make(map[string]attributeStrings)
	pk := primaryKeys(typeOf)[0]

	id := utils.ManyToManyNamePattern(pk.Name, nameTypeOf)
	mtm.ids[utils.TableNamePattern(nameTypeOf)] = createAttributeStrings(table, id)

	// target id
	pkTarget := primaryKeys(targetTypeOf)[0]
	id = utils.ManyToManyNamePattern(pkTarget.Name, nameTargetTypeOf)

	mtm.ids[utils.TableNamePattern(nameTargetTypeOf)] = createAttributeStrings(table, id)
	return mtm
}

type manyToOne struct {
	pk                  *pk
	targetTable         string
	id                  string
	attributeName       string
	structAttributeName string
	targetPkName        string
	hasMany             bool
}

func createManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, hasMany bool) *manyToOne {
	mto := new(manyToOne)
	targetPkName := primaryKeys(typeOf)[0].Name
	mto.targetTable = utils.TableNamePattern(typeOf.Name())
	mto.id = fmt.Sprintf("%v.%v", utils.TableNamePattern(targetTypeOf.Name()), utils.ManyToOneNamePattern(targetPkName, typeOf.Name()))
	mto.hasMany = hasMany
	mto.attributeName = utils.ColumnNamePattern(primaryKeys(typeOf)[0].Name + typeOf.Name())
	mto.structAttributeName = typeOf.Name()
	mto.targetPkName = targetPkName
	return mto
}

type attributeStrings struct {
	selectName          string
	attributeName       string
	structAttributeName string
}

func createAttributeStrings(table string, attributeName string) attributeStrings {
	return attributeStrings{
		selectName:          fmt.Sprintf("%v.%v", table, utils.ColumnNamePattern(attributeName)),
		attributeName:       utils.ColumnNamePattern(attributeName),
		structAttributeName: attributeName,
	}
}

type pk struct {
	table         string
	autoIncrement bool
	fks           map[string]any
	attributeStrings
}

func createPk(table string, attributeName string, autoIncrement bool) *pk {
	table = utils.TableNamePattern(table)
	return &pk{
		table:            table,
		attributeStrings: createAttributeStrings(table, attributeName),
		autoIncrement:    autoIncrement,
		fks:              make(map[string]any)}
}

type att struct {
	attributeStrings
	pk *pk
}

func createAtt(attributeName string, pk *pk) *att {
	return &att{
		attributeStrings: createAttributeStrings(pk.table, attributeName), pk: pk}
}
