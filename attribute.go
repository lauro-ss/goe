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

	table = utils.TableNamePattern(table)
	mtm := new(manyToMany)
	mtm.table = table
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
	pk          *pk
	targetTable string
	attributeStrings
	targetPkName string
	hasMany      bool
}

func (m *manyToOne) getPrimaryKey() *pk {
	return m.pk
}

func createManyToOne(typeOf reflect.Type, targetTypeOf reflect.Type, hasMany bool) *manyToOne {
	mto := new(manyToOne)
	targetPkName := primaryKeys(typeOf)[0].Name
	mto.targetTable = utils.TableNamePattern(typeOf.Name())
	mto.selectName = fmt.Sprintf("%v.%v", utils.TableNamePattern(targetTypeOf.Name()), utils.ManyToOneNamePattern(targetPkName, typeOf.Name()))
	mto.hasMany = hasMany
	mto.attributeName = utils.ColumnNamePattern(utils.ManyToOneNamePattern(targetPkName, typeOf.Name()))
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

func (p *pk) getPrimaryKey() *pk {
	return p
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

func (a *att) getPrimaryKey() *pk {
	return a.pk
}

func createAtt(attributeName string, pk *pk) *att {
	return &att{
		attributeStrings: createAttributeStrings(pk.table, attributeName), pk: pk}
}

func (p *pk) buildAttributeSelect(b *builder) {
	b.sql.WriteString(p.selectName)
	b.structColumns = append(b.structColumns, p.structAttributeName)
	b.pks.add(p)
}

func (a *att) buildAttributeSelect(b *builder) {
	b.sql.WriteString(a.selectName)
	b.structColumns = append(b.structColumns, a.structAttributeName)
	b.pks.add(a.pk)
}

func (m *manyToOne) buildAttributeSelect(b *builder) {
	b.sql.WriteString(m.selectName)
	b.structColumns = append(b.structColumns, m.structAttributeName)
	b.pks.add(m.pk)
}

func (p *pk) buildAttributeInsert(b *builder) {
	if !p.autoIncrement {
		b.sql.WriteString(p.attributeName)
		b.attrNames = append(b.attrNames, p.structAttributeName)
	}
}

func (a *att) buildAttributeInsert(b *builder) {
	b.sql.WriteString(a.attributeName)
	b.attrNames = append(b.attrNames, a.structAttributeName)
}

func (m *manyToOne) buildAttributeInsert(b *builder) {
	b.sql.WriteString(m.attributeName)
	b.attrNames = append(b.attrNames, m.structAttributeName)
	b.targetFksNames[m.structAttributeName] = m.targetPkName
}

func (p *pk) buildAttributeUpdate(b *builder) {
	if !p.autoIncrement {
		b.attrNames = append(b.attrNames, p.attributeName)
		b.structColumns = append(b.structColumns, p.structAttributeName)
	}
	b.pks.add(p)
}

func (a *att) buildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, a.attributeName)
	b.structColumns = append(b.structColumns, a.structAttributeName)
}

func (m *manyToOne) buildAttributeUpdate(b *builder) {
	b.attrNames = append(b.attrNames, m.attributeName)
	b.structColumns = append(b.structColumns, m.structAttributeName)
	b.targetFksNames[m.structAttributeName] = m.targetPkName
}

func (p *pk) buildComplexOperator(o string, v any) operator {
	return createComplexOperator(p.selectName, o, v, p)
}

func (a *att) buildComplexOperator(o string, v any) operator {
	return createComplexOperator(a.selectName, o, v, a.pk)
}

func (m *manyToOne) buildComplexOperator(o string, v any) operator {
	return createComplexOperator(m.selectName, o, v, m.pk)
}
