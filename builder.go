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
	sql            *strings.Builder
	args           []string
	argsAny        []any
	structColumns  []string          //select and update
	attrNames      []string          //insert and update
	targetFksNames map[string]string //insert and update
	brs            []operator
	queue          *statementQueue
	tables         *statementQueue
	pks            *pkQueue
	queryType      int8
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

func (b *builder) buildSelect(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementSELECT)

	b.structColumns = make([]string, 0, len(b.args))

	f := addrMap[b.args[0]]
	b.tables.add(createStatement(f.getPrimaryKey().table, writeTABLE))
	f.buildAttributeSelect(b)

	for _, v := range b.args[1:] {
		addrMap[v].buildAttributeSelect(b)
	}

	b.queue.add(statementFROM)
}

func (b *builder) buildSelectJoins(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementSELECT)

	f := addrMap[b.args[0]]
	b.tables.add(createStatement(f.getPrimaryKey().table, writeTABLE))
	b.pks.add(f.getPrimaryKey())

	for _, v := range b.args[1:] {
		f = addrMap[v]
		b.tables.add(createStatement(f.getPrimaryKey().table, writeTABLE))
		b.pks.add(f.getPrimaryKey())
	}

	b.queue.add(statementFROM)
}

func (b *builder) buildSqlSelect() {
	b.buildTables()
	b.buildWhere()
	writeSelect(b.sql, b.queue)
}

func (b *builder) buildSqlInsert() {
	writeInsert(b.sql, b.queue)
}

func (b *builder) buildSqlUpdate() {
	b.buildWhere()
	writeUpdate(b.sql, b.queue)
}

func (b *builder) buildSqlDelete() {
	b.buildWhere()
	writeDelete(b.sql, b.queue)
}

func (b *builder) buildSqlUpdateIn() {
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
			st := buildWhereIn(b.pks, v.pk, argsCount, v)
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
						createStatement(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pk.selectName, fk.selectName), writeJOIN),
					)
				} else {
					stQueue.add(
						createStatement(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pks.findPk(table.keyword).selectName, fk.selectName), writeJOIN),
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

func (b *builder) buildInsert(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementINSERT)

	b.queue.add(statementINTO)

	b.targetFksNames = make(map[string]string)
	b.attrNames = make([]string, 0, len(b.args))

	f := addrMap[b.args[0]]
	b.queue.add(createStatement(f.getPrimaryKey().table, writeDML))
	b.pks.add(f.getPrimaryKey())
	f.buildAttributeInsert(b)

	for _, v := range b.args[1:] {
		addrMap[v].buildAttributeInsert(b)
	}

	b.queue.add(statementVALUES)
}

func (b *builder) buildInsertIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementINSERT)

	b.queue.add(statementINTO)

	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, writeTABLE))
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
	b.queue.add(createStatement(mtmValue.table, writeDML))
	b.queue.add(createStatement(mtmValue.ids[pk1.table].attributeName, writeATT))
	b.queue.add(createStatement(mtmValue.ids[pk2.table].attributeName, writeATT))
	b.queue.add(statementVALUES)

	b.queue.add(createStatement("$1", writeATT))
	b.queue.add(createStatement("$2", writeATT))
}

func (b *builder) buildUpdate(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementUPDATE)

	b.targetFksNames = make(map[string]string)
	b.structColumns = make([]string, 0, len(b.args))
	b.attrNames = make([]string, 0, len(b.args))

	b.queue.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, writeDML))
	b.queue.add(statementSET)
	addrMap[b.args[0]].buildAttributeUpdate(b)

	for _, v := range b.args[1:] {
		addrMap[v].buildAttributeUpdate(b)
	}
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

func (b *builder) buildUpdateIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementUPDATE)

	b.attrNames = make([]string, 0, len(b.args))
	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, writeTABLE))
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
		b.queue.add(createStatement(mtmValue.table, writeDML))
		b.queue.add(statementSET)
		b.queue.add(createStatement(fmt.Sprintf("%v = $1", mtmValue.ids[pk2.table].attributeName), writeATT))
	}
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

func (b *builder) buildDelete(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementDELETE)
	b.queue.add(statementFROM)
	b.queue.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, writeDML))
}

func (b *builder) buildDeleteIn(addrMap map[string]field) {
	//TODO: Set a drive type to share stm
	b.queue.add(statementDELETE)
	b.queue.add(statementFROM)

	b.tables.add(createStatement(addrMap[b.args[0]].getPrimaryKey().table, writeTABLE))
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
		b.queue.add(createStatement(mtmValue.table, writeDML))
		b.buildWhereIn()

		writeDelete(b.sql, b.queue)
	}
	//TODO: add error
}
