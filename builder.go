package goe

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	querySELECT int8 = 1
	queryINSERT int8 = 2
	queryUPDATE int8 = 3
	queryDELETE int8 = 4
)

var (
	// Select keywords
	statementSELECT = &statement{
		keyword: "SELECT",
		tip:     writeDML,
	}
	statementFROM = &statement{
		keyword: "FROM",
		tip:     writeDML,
	}
	statementWHERE = &statement{
		keyword: "WHERE",
		tip:     writeMIDDLE,
	}

	// Insert keywords
	statementINSERT = &statement{
		keyword: "INSERT",
		tip:     writeDML,
	}
	statementINTO = &statement{
		keyword: "INTO",
		tip:     writeDML,
	}
	statementVALUES = &statement{
		keyword: "VALUES",
		tip:     writeDML,
	}
	statementRETURNING = &statement{
		keyword: "RETURNING",
		tip:     writeDML,
	}

	// Update keywords
	statementUPDATE = &statement{
		keyword: "UPDATE",
		tip:     writeDML,
	}
	statementSET = &statement{
		keyword: "SET",
		tip:     writeDML,
	}

	statementDELETE = &statement{
		keyword: "DELETE",
		tip:     writeDML,
	}
)

type builder struct {
	sql       *strings.Builder
	args      []string
	argsAny   []any
	attrNames []string
	brs       []operator
	queue     *statementQueue
	tables    *statementQueue
	pks       *pkQueue
	queryType int8
}

func createBuilder(qt int8) *builder {
	return &builder{
		sql:       &strings.Builder{},
		queue:     createStatementQueue(),
		tables:    createStatementQueue(),
		queryType: qt,
		pks:       createPkQueue()}
}

type statement struct {
	keyword     string
	allowCopies bool
	tip         int8
}

func createStatement(k string, t int8) *statement {
	return &statement{keyword: k, tip: t}
}

func (b *builder) buildSelect(addrMap map[string]any) (structColumns []string) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementSELECT)

	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.queue.add(createStatement(atr.selectName, writeATT))
			b.tables.add(createStatement(atr.pk.table, writeTABLE))
			structColumns = append(structColumns, atr.structAttributeName)
			//TODO: Add a list pk?
			b.pks.add(atr.pk)
		case *pk:
			b.queue.add(createStatement(atr.selectName, writeATT))
			b.tables.add(createStatement(atr.table, writeTABLE))
			structColumns = append(structColumns, atr.structAttributeName)
			//TODO: Add a list pk?
			b.pks.add(atr)
		}
	}

	b.queue.add(statementFROM)
	return structColumns
}

func (b *builder) buildSelectJoins(addrMap map[string]any) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementSELECT)

	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.tables.add(createStatement(atr.pk.table, writeTABLE))

			//TODO: Add a list pk?
			b.pks.add(atr.pk)
		case *pk:
			b.tables.add(createStatement(atr.table, writeTABLE))

			//TODO: Add a list pk?
			b.pks.add(atr)
		}
	}

	b.queue.add(statementFROM)
}

func (b *builder) buildSql() {
	//TODO: Remove this switch
	switch b.queryType {
	case querySELECT:
		b.buildTables()
		b.buildWhere()
		writeSelect(b.sql, b.queue)
	case queryINSERT:
		writeInsert(b.sql, b.queue)
	case queryUPDATE:
		b.buildWhere()
		writeUpdate(b.sql, b.queue)
	case queryDELETE:
		break
	}
}

func (b *builder) buildSqlDelete() {
	b.buildWhere()
	writeDelete(b.sql, b.queue)
}

func (b *builder) buildeSqlUpdateIn() {
	b.buildWhereIn()
	writeUpdate(b.sql, b.queue)
}

func (b *builder) buildWhere() {
	if len(b.brs) == 0 {
		return
	}
	b.queue.add(statementWHERE)
	argsCount := len(b.argsAny) + 1
	for _, op := range b.brs {
		switch v := op.(type) {
		case complexOperator:
			v.setValueFlag(fmt.Sprintf("$%v", argsCount))
			b.queue.add(createStatement(v.operation(), 0))
			b.argsAny = append(b.argsAny, v.value)
			argsCount++
		case simpleOperator:
			b.queue.add(createStatement(v.operation(), 0))
		}
	}
}

func (b *builder) buildWhereIn() {
	if len(b.brs) == 0 {
		return
	}
	b.queue.add(statementWHERE)
	argsCount := len(b.argsAny) + 1

	for _, op := range b.brs {
		switch v := op.(type) {
		case complexOperator:
			st := buildWhereIn(b.pks, v.pk, argsCount)
			if st != nil {
				b.queue.add(st)
				b.argsAny = append(b.argsAny, v.value)
				argsCount++
			}
		case simpleOperator:
			b.queue.add(createStatement(v.operation(), 0))
		}
	}
}

func buildWhereIn(pkQueue *pkQueue, brPk *pk, argsCount int) *statement {
	pk2 := pkQueue.get()
	if pk2 == nil {
		pk2 = pkQueue.get()
	}
	for pk2 != nil {
		mtm := brPk.fks[pk2.table]
		if mtm != nil {
			mtmValue := mtm.(*manyToMany)
			return createStatement(fmt.Sprintf("%v = $%v", mtmValue.ids[brPk.table].attributeName, argsCount), 0)
		}
		pk2 = pkQueue.get()
	}
	return nil
}

func (b *builder) buildTables() {
	b.queue.add(b.tables.get())
	if b.tables.size >= 1 {
		for table := b.tables.get(); table != nil; {
			buildJoins(table, b.pks, b.queue)
			table = b.tables.get()
		}
	}
}

func buildJoins(table *statement, pks *pkQueue, stQueue *statementQueue) {
	var skipTable string
	for pk := pks.get(); pk != nil; {
		if pk.table != table.keyword && skipTable != table.keyword {
			switch fk := pk.fks[table.keyword].(type) {
			case *manyToOne:
				if fk.hasMany {
					stQueue.add(
						createStatement(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pk.selectName, fk.id), writeJOIN),
					)
				} else {
					stQueue.add(
						createStatement(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pks.findPk(table.keyword).selectName, fk.id), writeJOIN),
					)
				}
				// skips the table keyword that has already be matched
				skipTable = table.keyword
			case *manyToMany:
				stQueue.add(
					createStatement(fmt.Sprintf("inner join %v on (%v = %v)", fk.table, pk.selectName, fk.ids[pk.table].selectName), writeJOIN),
				)
				stQueue.add(
					createStatement(
						fmt.Sprintf(
							"inner join %v on (%v = %v)",
							table.keyword, fk.ids[table.keyword].selectName,
							pks.findPk(table.keyword).selectName), writeJOIN,
					),
				)
				// skips the table keyword that has already be matched
				skipTable = table.keyword
			}
		}
		pk = pks.get()
	}
}

func (b *builder) buildUpdate(addrMap map[string]any) (targetFksNames map[string]string, strNames []string) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementUPDATE)

	targetFksNames = make(map[string]string)
	strNames = make([]string, 0, len(b.args))
	attrNames := make([]string, 0, len(b.args))
	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.queue.add(createStatement(atr.pk.table, writeDML))
			b.queue.add(statementSET)
			//b.queue.add(createStatement(fmt.Sprintf("%v = %v", atr.attributeName, "$1"), writeATT))
			attrNames = append(attrNames, atr.attributeName)
			strNames = append(strNames, atr.structAttributeName)
		case *pk:
			if !atr.autoIncrement {
				b.queue.add(createStatement(atr.table, writeDML))
				b.queue.add(statementSET)
				//b.queue.add(createStatement(fmt.Sprintf("%v = %v", atr.attributeName, "$1"), writeATT))
				attrNames = append(attrNames, atr.attributeName)
				strNames = append(strNames, atr.structAttributeName)
			}
			b.pks.add(atr)
		case *manyToOne:
			b.queue.add(createStatement(atr.pk.table, writeDML))
			b.queue.add(statementSET)
			attrNames = append(attrNames, atr.attributeName)
			strNames = append(strNames, atr.structAttributeName)
			targetFksNames[atr.structAttributeName] = atr.targetPkName
		}
	}
	b.attrNames = attrNames
	return targetFksNames, strNames
}

func (b *builder) buildUpdateIn(addrMap map[string]any) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementUPDATE)

	attrNames := make([]string, 0, len(b.args))
	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.tables.add(createStatement(atr.pk.table, writeTABLE))
			b.pks.add(atr.pk)
		case *pk:
			b.tables.add(createStatement(atr.table, writeTABLE))
			b.pks.add(atr)
		}
	}
	b.attrNames = attrNames
}

func (b *builder) buildInsert(addrMap map[string]any) map[string]string {
	//TODO: Set a drive type to share stm
	b.queue.add(statementINSERT)

	b.queue.add(statementINTO)

	targetFksNames := make(map[string]string)
	attrNames := make([]string, 0, len(b.args))
	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.queue.add(createStatement(atr.pk.table, writeDML))
			b.queue.add(createStatement(atr.attributeName, writeATT))
			attrNames = append(attrNames, atr.structAttributeName)

		case *pk:
			if !atr.autoIncrement {
				b.queue.add(createStatement(atr.table, writeDML))
				b.queue.add(createStatement(atr.attributeName, writeATT))
				attrNames = append(attrNames, atr.structAttributeName)
			}
			b.pks.add(atr)
		case *manyToOne:
			b.queue.add(createStatement(atr.pk.table, writeDML))
			b.queue.add(createStatement(atr.attributeName, writeATT))
			attrNames = append(attrNames, atr.structAttributeName)
			targetFksNames[atr.structAttributeName] = atr.targetPkName
		}
	}

	b.attrNames = attrNames

	b.queue.add(statementVALUES)

	return targetFksNames
}

func (b *builder) buildInsertIn(addrMap map[string]any) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementINSERT)

	b.queue.add(statementINTO)

	attrNames := make([]string, 0, len(b.args))
	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.tables.add(createStatement(atr.pk.table, writeTABLE))
			b.pks.add(atr.pk)
		case *pk:
			b.tables.add(createStatement(atr.table, writeTABLE))
			b.pks.add(atr)
		}
	}

	b.attrNames = attrNames
}

func (b *builder) buildSet(value reflect.Value, targetFksNames map[string]string, strNames []string) {
	var valueField reflect.Value
	b.argsAny = make([]any, 0, len(b.attrNames))
	c := 1
	for i, attr := range b.attrNames {
		valueField = value.FieldByName(strNames[i])
		switch valueField.Kind() {
		case reflect.Struct:
			if valueField.Type().Name() == "Time" {
				b.queue.add(createStatement(fmt.Sprintf("%v = $%v", attr, c), writeATT))
				b.argsAny = append(b.argsAny, valueField.Interface())
				c++
				continue
			}
			if !valueField.FieldByName(targetFksNames[strNames[i]]).IsZero() {
				b.queue.add(createStatement(fmt.Sprintf("%v = $%v", attr, c), writeATT))
				b.argsAny = append(b.argsAny, valueField.FieldByName(targetFksNames[strNames[i]]).Interface())
				c++
			}
			continue
		case reflect.Pointer:
			if !valueField.IsNil() && valueField.Elem().Kind() == reflect.Struct {
				b.queue.add(createStatement(fmt.Sprintf("%v = $%v", attr, c), writeATT))
				b.argsAny = append(b.argsAny, valueField.Elem().FieldByName(targetFksNames[strNames[i]]).Interface())
				c++
				continue
			}
		}
		b.queue.add(createStatement(fmt.Sprintf("%v = $%v", attr, c), writeATT))
		b.argsAny = append(b.argsAny, valueField.Interface())
		c++
	}
}

func (b *builder) buildSetIn() {
	if b.tables.size != 2 {
		return
	}

	stTable := b.tables.get()

	// skips the first primary key
	b.pks.get()
	pk2 := b.pks.get()

	mtm := pk2.fks[stTable.keyword]
	if mtm == nil {
		return
	}

	mtmValue := mtm.(*manyToMany)
	b.queue.add(createStatement(mtmValue.table, writeDML))
	b.queue.add(statementSET)
	b.queue.add(createStatement(fmt.Sprintf("%v = $1", mtmValue.ids[pk2.table].attributeName), writeATT))
}

func (b *builder) buildValues(value reflect.Value, targetFksNames map[string]string) string {
	var valueField reflect.Value
	b.argsAny = make([]any, 0, len(b.attrNames))
	for i, attr := range b.attrNames {
		b.queue.add(createStatement(fmt.Sprintf("$%v", i+1), writeATT))
		valueField = value.FieldByName(attr)
		switch valueField.Kind() {
		case reflect.Struct:
			if valueField.Type().Name() != "Time" {
				b.argsAny = append(b.argsAny, valueField.FieldByName(targetFksNames[attr]).Interface())
				continue
			}
		case reflect.Pointer:
			if !valueField.IsNil() && valueField.Elem().Kind() == reflect.Struct {
				b.argsAny = append(b.argsAny, valueField.Elem().FieldByName(targetFksNames[attr]).Interface())
				continue
			}
		}
		b.argsAny = append(b.argsAny, valueField.Interface())
	}
	pk := b.pks.get()
	b.queue.add(statementRETURNING)
	st := createStatement(pk.attributeName, 0)
	st.allowCopies = true
	b.queue.add(st)
	return pk.structAttributeName

}

func (b *builder) buildValuesIn() {
	if b.tables.size != 2 {
		return
	}
	stTable := b.tables.get()

	pk1 := b.pks.get()
	pk2 := b.pks.get()

	mtm := pk2.fks[stTable.keyword]
	if mtm == nil {
		return
	}

	mtmValue := mtm.(*manyToMany)
	b.queue.add(createStatement(mtmValue.table, writeDML))
	b.queue.add(createStatement(mtmValue.ids[pk1.table].attributeName, writeATT))
	b.queue.add(createStatement(mtmValue.ids[pk2.table].attributeName, writeATT))
	b.queue.add(statementVALUES)

	b.queue.add(createStatement("$1", writeATT))
	b.queue.add(createStatement("$2", writeATT))
}

func (b *builder) buildDelete(addrMap map[string]any) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementDELETE)
	b.queue.add(statementFROM)

	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.queue.add(createStatement(atr.pk.table, writeDML))

		case *pk:
			b.queue.add(createStatement(atr.table, writeDML))
		}
	}

}

func (b *builder) buildDeleteIn(addrMap map[string]any) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementDELETE)
	b.queue.add(statementFROM)

	//TODO Better Query
	for _, v := range b.args {
		switch atr := addrMap[v].(type) {
		case *att:
			b.pks.add(atr.pk)
			b.tables.add(createStatement(atr.pk.table, writeTABLE))
		case *pk:
			b.pks.add(atr)
			b.tables.add(createStatement(atr.table, writeTABLE))
		}
	}

}

func (b *builder) buildSqlDeleteIn() {
	if b.tables.size != 2 {
		return
	}
	stTable := b.tables.get()

	pk1 := b.pks.get()
	pk2 := b.pks.get()

	mtm := pk2.fks[stTable.keyword]
	if mtm == nil {
		return
	}

	mtmValue := mtm.(*manyToMany)
	b.queue.add(createStatement(mtmValue.table, writeDML))
	if len(b.argsAny) != 0 {
		b.queue.add(createStatement(fmt.Sprintf("WHERE %v = $1", mtmValue.ids[pk1.table].attributeName), writeATT))
		if len(b.argsAny) == 2 {
			b.queue.add(createStatement(fmt.Sprintf(" AND %v = $2", mtmValue.ids[pk2.table].attributeName), writeATT))
		}
	}

	writeDelete(b.sql, b.queue)
}
