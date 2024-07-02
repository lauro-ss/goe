package goe

import (
	"fmt"
	"reflect"
	"strings"
)

type builder struct {
	sql            *strings.Builder
	args           []string
	argsAny        []any
	structColumns  []string          //select and update
	attrNames      []string          //insert and update
	targetFksNames map[string]string //insert and update
	brs            []operator
	tables         *statementQueue
	pks            *pkQueue
}

func createBuilder() *builder {
	return &builder{
		sql:    &strings.Builder{},     //TODO: Add grow for sql builder
		tables: createStatementQueue(), //TODO: Change to string queue
		pks:    createPkQueue()}        //TODO: Change to string queue
}

type statement struct {
	keyword     string
	allowCopies bool
	tip         int8
}

func createStatement(k string, t int8) *statement {
	return &statement{keyword: k, tip: t}
}

func (b *builder) buildSelect(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("SELECT")
	b.sql.WriteRune(' ')

	b.structColumns = make([]string, 0, len(b.args))

	for _, v := range b.args[1:] {
		addrMap[v].buildAttributeSelect(b)
		b.sql.WriteRune(',')
	}
	addrMap[b.args[0]].buildAttributeSelect(b)
	b.sql.WriteString(" FROM ")
	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, 0))
}

func (b *builder) buildSelectJoins(addrMap map[string]field) {
	for _, v := range b.args[1:] {
		b.tables.add(createStatement(addrMap[v].getPrimaryKey().table, 0))
		b.pks.add(addrMap[v].getPrimaryKey())
	}

	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, 0))
	b.pks.add(addrMap[b.args[0]].getPrimaryKey())
}

func (b *builder) buildSqlSelect() {
	b.buildTables()
	b.buildWhere()
	b.sql.WriteRune(';')
}

func (b *builder) buildSqlUpdate() {
	b.buildWhere()
	b.sql.WriteRune(';')
}

func (b *builder) buildSqlDelete() {
	b.buildWhere()
	b.sql.WriteRune(';')
}

func (b *builder) buildSqlUpdateIn() {
	b.buildWhereIn()
	b.sql.WriteRune(';')
}

func (b *builder) buildWhere() {
	if len(b.brs) == 0 {
		return
	}
	b.sql.WriteRune('\n')
	b.sql.WriteString("WHERE ")
	argsCount := len(b.argsAny) + 1
	for _, op := range b.brs {
		switch v := op.(type) {
		case complexOperator:
			v.setValueFlag(fmt.Sprintf("$%v", argsCount))
			b.sql.WriteString(v.operation())
			b.argsAny = append(b.argsAny, v.value)
			argsCount++
		case simpleOperator:
			b.sql.WriteString(v.operation())
		}
	}
}

func (b *builder) buildWhereIn() {
	if len(b.brs) == 0 {
		return
	}
	b.sql.WriteRune('\n')
	b.sql.WriteString("WHERE ")
	argsCount := len(b.argsAny) + 1

	for _, op := range b.brs {
		switch v := op.(type) {
		case complexOperator:
			st := buildWhereIn(b.pks, v.pk, argsCount, v)
			if st != nil {
				b.sql.WriteString(st.keyword)
				b.argsAny = append(b.argsAny, v.value)
				argsCount++
			}
		case simpleOperator:
			b.sql.WriteString(v.operation())
		}
	}
}

func buildWhereIn(pkQueue *pkQueue, brPk *pk, argsCount int, v complexOperator) *statement {
	pk2 := pkQueue.get()
	if pk2 == nil {
		pk2 = pkQueue.get()
	}
	for pk2 != nil {
		mtm := brPk.fks[pk2.table]
		if mtm != nil {
			if mtmValue, ok := mtm.(*manyToMany); ok {
				v.setValueFlag(fmt.Sprintf("$%v", argsCount))
				v.setArgument(mtmValue.ids[brPk.table].attributeName)
				return createStatement(v.operation(), 0)
			}
		}
		pk2 = pkQueue.get()
	}
	return nil
}

func (b *builder) buildTables() {
	b.sql.WriteString(b.tables.get().keyword)
	if b.tables.size >= 1 {
		for table := b.tables.get(); table != nil; {
			buildJoins(table, b.pks, b.sql)
			table = b.tables.get()
		}
	}
}

func buildJoins(table *statement, pks *pkQueue, sql *strings.Builder) {
	var skipTable string
	for pk := pks.get(); pk != nil; {
		if pk.table != table.keyword && skipTable != table.keyword {
			switch fk := pk.fks[table.keyword].(type) {
			case *manyToOne:
				if fk.hasMany {
					sql.WriteRune('\n')
					sql.WriteString(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pk.selectName, fk.selectName))
				} else {
					sql.WriteRune('\n')
					sql.WriteString(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pk.selectName, fk.selectName))
				}
				// skips the table keyword that has already be matched
				skipTable = table.keyword
			case *manyToMany:
				sql.WriteRune('\n')
				sql.WriteString(fmt.Sprintf("inner join %v on (%v = %v)", fk.table, pk.selectName, fk.ids[pk.table].selectName))
				sql.WriteRune('\n')
				sql.WriteString(fmt.Sprintf(
					"inner join %v on (%v = %v)",
					table.keyword, fk.ids[table.keyword].selectName,
					pks.findPk(table.keyword).selectName))
				// skips the table keyword that has already be matched
				skipTable = table.keyword
			}
		}
		pk = pks.get()
	}
}

func (b *builder) buildInsert(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("INSERT ")
	b.sql.WriteString("INTO ")

	b.targetFksNames = make(map[string]string)
	b.attrNames = make([]string, 0, len(b.args))

	f := addrMap[b.args[0]]
	b.sql.WriteString(f.getPrimaryKey().table)
	b.sql.WriteString(" (")
	b.pks.add(f.getPrimaryKey())
	f.buildAttributeInsert(b)

	for _, v := range b.args[1:] {
		b.sql.WriteRune(',')
		addrMap[v].buildAttributeInsert(b)
	}
	b.sql.WriteString(") ")
	b.sql.WriteString("VALUES ")
}

func (b *builder) buildInsertIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("INSERT ")
	b.sql.WriteString("INTO ")

	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, 0))
	b.pks.add(addrMap[b.args[0]].getPrimaryKey())
	b.pks.add(addrMap[b.args[1]].getPrimaryKey())
}

func (b *builder) buildValuesIn() {
	stTable := b.tables.get()

	pk1 := b.pks.get()
	pk2 := b.pks.get()

	mtm := pk2.fks[stTable.keyword]
	if mtm == nil {
		return
	}

	mtmValue := mtm.(*manyToMany)
	b.sql.WriteString(mtmValue.table)
	b.sql.WriteString(" (")
	b.sql.WriteString(mtmValue.ids[pk1.table].attributeName)
	b.sql.WriteString(",")
	b.sql.WriteString(mtmValue.ids[pk2.table].attributeName)
	b.sql.WriteString(") ")
	b.sql.WriteString("VALUES ")
	b.sql.WriteString("($1,$2);")
}

func (b *builder) buildUpdate(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("UPDATE ")

	b.targetFksNames = make(map[string]string)
	b.structColumns = make([]string, 0, len(b.args))
	b.attrNames = make([]string, 0, len(b.args))

	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
	b.sql.WriteString(" SET ")
	addrMap[b.args[0]].buildAttributeUpdate(b)

	for _, v := range b.args[1:] {
		addrMap[v].buildAttributeUpdate(b)
	}
}

func (b *builder) buildSet(value reflect.Value) {
	b.argsAny = make([]any, 0, len(b.attrNames))
	var c uint16 = 1
	buildSetField(value.FieldByName(b.structColumns[0]), b.attrNames[0], b, c)
	for i, attr := range b.attrNames[1:] {
		b.sql.WriteRune(',')
		c++
		buildSetField(value.FieldByName(b.structColumns[i+1]), attr, b, c)
	}
}

func buildSetField(valueField reflect.Value, fieldName string, b *builder, c uint16) {
	switch valueField.Kind() {
	case reflect.Struct:
		if valueField.Type().Name() == "Time" {
			b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
			b.argsAny = append(b.argsAny, valueField.Interface())
			c++
			return
		}
		if !valueField.FieldByName(b.targetFksNames[fieldName]).IsZero() {
			b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
			b.argsAny = append(b.argsAny, valueField.FieldByName(b.targetFksNames[fieldName]).Interface())
			c++
		}
		return
	case reflect.Pointer:
		if !valueField.IsNil() && valueField.Elem().Kind() == reflect.Struct {
			b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
			b.argsAny = append(b.argsAny, valueField.Elem().FieldByName(b.targetFksNames[fieldName]).Interface())
			c++
			return
		}
	}
	b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
	b.argsAny = append(b.argsAny, valueField.Interface())
	c++
}

func (b *builder) buildUpdateIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("UPDATE ")

	b.attrNames = make([]string, 0, len(b.args))
	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, 0))
	b.pks.add(addrMap[b.args[0]].getPrimaryKey())
	b.pks.add(addrMap[b.args[1]].getPrimaryKey())
}

func (b *builder) buildSetIn() {

	stTable := b.tables.get()

	// skips the first primary key
	b.pks.get()
	pk2 := b.pks.get()

	mtm := pk2.fks[stTable.keyword]
	if mtm == nil {
		return
	}

	if mtmValue, ok := mtm.(*manyToMany); ok {
		b.sql.WriteString(mtmValue.table)
		b.sql.WriteString(" SET ")
		b.sql.WriteString(fmt.Sprintf("%v = $1", mtmValue.ids[pk2.table].attributeName))
	}
}

func (b *builder) buildValues(value reflect.Value) string {
	b.sql.WriteString("(")
	b.argsAny = make([]any, 0, len(b.attrNames))

	c := 2
	b.sql.WriteString("$1")
	buildValueField(value.FieldByName(b.attrNames[0]), b.attrNames[0], b)
	for _, attr := range b.attrNames[1:] {
		b.sql.WriteRune(',')
		b.sql.WriteString(fmt.Sprintf("$%v", c))
		buildValueField(value.FieldByName(attr), attr, b)
		c++
	}
	pk := b.pks.get()
	b.sql.WriteString(") ")
	b.sql.WriteString("RETURNING ")
	st := createStatement(pk.attributeName, 0)
	st.allowCopies = true
	b.sql.WriteString(pk.attributeName)
	b.sql.WriteRune(';')
	return pk.structAttributeName

}

func buildValueField(valueField reflect.Value, fieldName string, b *builder) {
	switch valueField.Kind() {
	case reflect.Struct:
		if valueField.Type().Name() != "Time" {
			b.argsAny = append(b.argsAny, valueField.FieldByName(b.targetFksNames[fieldName]).Interface())
			return
		}
	case reflect.Pointer:
		if !valueField.IsNil() && valueField.Elem().Kind() == reflect.Struct {
			b.argsAny = append(b.argsAny, valueField.Elem().FieldByName(b.targetFksNames[fieldName]).Interface())
			return
		}
	}
	b.argsAny = append(b.argsAny, valueField.Interface())
}

func (b *builder) buildDelete(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("DELETE FROM ")
	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
}

func (b *builder) buildDeleteIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("DELETE FROM ")

	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, 0))
	b.pks.add(addrMap[b.args[0]].getPrimaryKey())
	b.pks.add(addrMap[b.args[1]].getPrimaryKey())
}

func (b *builder) buildSqlDeleteIn() {
	stTable := b.tables.get()

	b.pks.get()
	pk2 := b.pks.get()

	mtm := pk2.fks[stTable.keyword]
	if mtm == nil {
		//TODO: add error
		return
	}

	if mtmValue, ok := mtm.(*manyToMany); ok {
		b.sql.WriteString(mtmValue.table)
		b.buildWhereIn()
		b.sql.WriteRune(';')
	}
	//TODO: add error
}
