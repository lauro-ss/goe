package goe

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var ErrInvalidWhere = errors.New("goe: invalid where operation. try sending a pointer as parameter")
var ErrNoMatchesTables = errors.New("don't have any many to one or many to many relationship")
var ErrNotManyToMany = errors.New("don't have a many to many relationship")

type builder struct {
	sql           *strings.Builder
	args          []uintptr
	aggregates    []aggregate
	argsAny       []any
	structColumns []string //select and update
	attrNames     []string //insert and update
	orderBy       string
	limit         uint
	offset        uint
	joins         []string
	brs           []operator
	table         string
	tablesPk      []*pk
}

func createBuilder() *builder {
	return &builder{
		sql: &strings.Builder{}}
}

func (b *builder) buildSelect(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("SELECT ")

	if len(b.aggregates) > 0 {
		b.buildAggregates()
	}

	lenArgs := len(b.args)
	if lenArgs == 0 {
		return
	}

	b.structColumns = make([]string, lenArgs)
	b.tablesPk = make([]*pk, 1)

	for i := range b.args[:lenArgs-1] {
		addrMap[b.args[i]].buildAttributeSelect(b, i)
		b.sql.WriteByte(44)
	}

	addrMap[b.args[lenArgs-1]].buildAttributeSelect(b, lenArgs-1)
	b.sql.WriteString(" FROM ")
	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
}

func (b *builder) buildAggregates() {
	for i := range b.aggregates[:len(b.aggregates)-1] {
		b.sql.WriteString(b.aggregates[i].String())
		b.sql.WriteByte(44)
	}
	b.sql.WriteString(b.aggregates[len(b.aggregates)-1].String())
	if len(b.args) == 0 {
		b.tablesPk = make([]*pk, 1)
		b.sql.WriteString(" FROM ")
		b.sql.WriteString(b.aggregates[0].field.getPrimaryKey().table)
		b.tablesPk[0] = b.aggregates[0].field.getPrimaryKey()
	}
}

func (b *builder) buildSelectJoins(addrMap map[uintptr]field, join string, argsJoins []uintptr) {
	b.tablesPk = append(b.tablesPk, make([]*pk, 2)...)
	c := len(b.tablesPk) - 2
	b.joins = append(b.joins, join)
	b.tablesPk[c] = addrMap[argsJoins[0]].getPrimaryKey()
	b.tablesPk[c+1] = addrMap[argsJoins[1]].getPrimaryKey()
}

func (b *builder) buildPage() {
	if b.limit != 0 {
		b.sql.WriteString(fmt.Sprintf(" LIMIT %v", b.limit))
	}
	if b.offset != 0 {
		b.sql.WriteString(fmt.Sprintf(" OFFSET %v", b.offset))
	}
}

func (b *builder) buildSqlSelect() (err error) {
	err = b.buildTables()
	if err != nil {
		return err
	}
	err = b.buildWhere()
	b.sql.WriteString(b.orderBy)
	b.buildPage()
	b.sql.WriteByte(59)
	return err
}

func (b *builder) buildSqlUpdate() (err error) {
	err = b.buildWhere()
	b.sql.WriteByte(59)
	return err
}

func (b *builder) buildSqlDelete() (err error) {
	err = b.buildWhere()
	b.sql.WriteByte(59)
	return err
}

func (b *builder) buildSqlUpdateIn() (err error) {
	err = b.buildWhereIn()
	b.sql.WriteByte(59)
	return err
}

func (b *builder) buildWhere() error {
	if len(b.brs) == 0 {
		return nil
	}
	b.sql.WriteByte(10)
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
		default:
			return ErrInvalidWhere
		}
	}
	return nil
}

func (b *builder) buildWhereIn() error {
	if len(b.brs) == 0 {
		return nil
	}
	b.sql.WriteByte(10)
	b.sql.WriteString("WHERE ")
	argsCount := len(b.argsAny) + 1

	for _, op := range b.brs {
		switch v := op.(type) {
		case complexOperator:
			st := buildWhereIn(b.tablesPk, v.pk, argsCount, v)
			if st != "" {
				b.sql.WriteString(st)
				b.argsAny = append(b.argsAny, v.value)
				argsCount++
			} else {
				return fmt.Errorf("goe: the tables %s and %s %w", b.table, v.pk.table, ErrNotManyToMany)
			}
		case simpleOperator:
			b.sql.WriteString(v.operation())
		default:
			return ErrInvalidWhere
		}
	}
	return nil
}

func buildWhereIn(pks []*pk, brPk *pk, argsCount int, v complexOperator) string {
	for i := range pks {
		mtm := brPk.fks[pks[i].table]
		if mtm != nil {
			if mtmValue, ok := mtm.(*manyToMany); ok {
				v.setValueFlag(fmt.Sprintf("$%v", argsCount))
				v.setArgument(mtmValue.ids[brPk.table].attributeName)
				return v.operation()
			}
		}
	}
	return ""
}

func (b *builder) buildTables() (err error) {
	if len(b.tablesPk) > 1 {
		c := 1
		for i := range b.joins {
			err = buildJoins(b.tablesPk[i+c], b.tablesPk[i+c+1], b.joins[i], b.sql, i+c+1, b.tablesPk)
			if err != nil {
				return err
			}
			c++
		}
	}
	return err
}

func buildJoins(pk1 *pk, pk2 *pk, join string, sql *strings.Builder, i int, pks []*pk) error {
	switch fk := pk1.fks[pk2.table].(type) {
	case *manyToOne:
		table := pk2.table
		for c := range pks[:i] {
			//switch table if pk2 is priority
			if pks[c].table == pk2.table {
				table = pk1.table
				break
			}
		}
		if fk.hasMany {
			sql.WriteByte(10)
			sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, table, pk1.selectName, fk.selectName))
		} else {
			sql.WriteByte(10)
			sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, table, pk2.selectName, fk.selectName))
		}
	case *manyToMany:
		pk1Value := *pk1
		pk2Value := *pk2
		for c := range pks[:i] {
			//switch pks if pk2 is priority
			if pks[c].table == pk2.table {
				pk1Value = *pks[c]
				pk2Value = *pk1
				break
			}
		}
		sql.WriteByte(10)
		sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, fk.table, pk1Value.selectName, fk.ids[pk1Value.table].selectName))
		sql.WriteByte(10)
		sql.WriteString(fmt.Sprintf(
			"%v %v on (%v = %v)",
			join,
			pk2Value.table, fk.ids[pk2Value.table].selectName,
			pk2Value.selectName))
	case *oneToOne:
		sql.WriteByte(10)
		var flag bool
		for c := range pks[:i] {
			if pks[c].table == pk2.table {
				flag = true
				break
			}
		}
		if flag {
			sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, pk1.table, pk2.selectName, fk.selectName))
			break
		}
		sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, pk2.table, pk2.selectName, fk.selectName))
	default:
		if fk = pk2.fks[pk1.table]; fk != nil {
			fk, ok := fk.(*oneToOne)
			if ok {
				sql.WriteByte(10)
				var flag bool
				for c := range pks[:i] {
					if pks[c].table == pk2.table {
						flag = true
						break
					}
				}
				if flag {
					sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, pk1.table, pk1.selectName, fk.selectName))
					break
				}
				sql.WriteString(fmt.Sprintf("%v %v on (%v = %v)", join, pk2.table, pk1.selectName, fk.selectName))
				return nil
			}
		}
		return fmt.Errorf("goe: the tables %s and %s %w", pk1.table, pk2.table, ErrNoMatchesTables)
	}
	return nil
}

func (b *builder) buildInsert(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("INSERT ")
	b.sql.WriteString("INTO ")

	b.attrNames = make([]string, 0, len(b.args))
	b.tablesPk = make([]*pk, 1)

	f := addrMap[b.args[0]]
	b.sql.WriteString(f.getPrimaryKey().table)
	b.sql.WriteString(" (")
	b.tablesPk[0] = f.getPrimaryKey()
	f.buildAttributeInsert(b)
	if !f.getPrimaryKey().autoIncrement {
		b.sql.WriteByte(44)
	}

	l := len(b.args[1:]) - 1

	a := b.args[1:]
	for i := range a {
		addrMap[a[i]].buildAttributeInsert(b)
		if i != l {
			b.sql.WriteByte(44)
		}
	}
	b.sql.WriteString(") ")
	b.sql.WriteString("VALUES ")
}

func (b *builder) buildValues(value reflect.Value) string {
	b.sql.WriteByte(40)
	b.argsAny = make([]any, 0, len(b.attrNames))

	c := 2
	b.sql.WriteString("$1")
	buildValueField(value.FieldByName(b.attrNames[0]), b)
	a := b.attrNames[1:]
	for i := range a {
		b.sql.WriteByte(44)
		b.sql.WriteString(fmt.Sprintf("$%v", c))
		buildValueField(value.FieldByName(a[i]), b)
		c++
	}
	pk := b.tablesPk[0]
	b.sql.Write([]byte{41, 32})
	b.sql.WriteString("RETURNING ")
	b.sql.WriteString(pk.attributeName)
	b.sql.WriteByte(59)
	return pk.structAttributeName

}

func (b *builder) buildBatchValues(value reflect.Value) string {
	b.argsAny = make([]any, 0, len(b.attrNames))

	c := 1
	buildBatchValues(value.Index(0), b, &c)
	c++
	for j := 1; j < value.Len(); j++ {
		b.sql.WriteByte(44)
		buildBatchValues(value.Index(j), b, &c)
		c++
	}
	pk := b.tablesPk[0]
	b.sql.WriteString(" RETURNING ")
	b.sql.WriteString(pk.attributeName)
	b.sql.WriteByte(59)
	return pk.structAttributeName

}

func buildBatchValues(value reflect.Value, b *builder, c *int) {
	b.sql.WriteByte(40)
	b.sql.WriteString(fmt.Sprintf("$%v", *c))
	buildValueField(value.FieldByName(b.attrNames[0]), b)
	a := b.attrNames[1:]
	for i := range a {
		b.sql.WriteByte(44)
		b.sql.WriteString(fmt.Sprintf("$%v", *c+1))
		buildValueField(value.FieldByName(a[i]), b)
		*c++
	}
	b.sql.WriteString(")")
}

func buildValueField(valueField reflect.Value, b *builder) {
	b.argsAny = append(b.argsAny, valueField.Interface())
}

func (b *builder) buildInsertIn(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("INSERT ")
	b.sql.WriteString("INTO ")

	b.tablesPk = make([]*pk, 2)

	b.table = addrMap[b.args[0]].getPrimaryKey().table
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
	b.tablesPk[1] = addrMap[b.args[1]].getPrimaryKey()
}

func (b *builder) buildValuesIn() error {
	pk1 := b.tablesPk[0]
	pk2 := b.tablesPk[1]

	mtm := pk2.fks[b.table]
	if mtm == nil {
		return fmt.Errorf("goe: the tables %s and %s %w", b.table, pk2.table, ErrNoMatchesTables)
	}

	mtmValue, ok := mtm.(*manyToMany)
	if !ok {
		return fmt.Errorf("goe: the tables %s and %s %w", b.table, pk2.table, ErrNotManyToMany)
	}
	b.sql.WriteString(mtmValue.table)
	b.sql.WriteString(" (")
	b.sql.WriteString(mtmValue.ids[pk1.table].attributeName)
	b.sql.WriteString(",")
	b.sql.WriteString(mtmValue.ids[pk2.table].attributeName)
	b.sql.WriteString(") ")
	b.sql.WriteString("VALUES ")
	b.sql.WriteString("($1,$2);")
	return nil
}

func (b *builder) buildValuesInBatch(v reflect.Value) error {
	pk1 := b.tablesPk[0]
	pk2 := b.tablesPk[1]

	mtm := pk2.fks[b.table]
	if mtm == nil {
		return fmt.Errorf("goe: the tables %s and %s %w", b.table, pk2.table, ErrNoMatchesTables)
	}

	mtmValue, ok := mtm.(*manyToMany)
	if !ok {
		return fmt.Errorf("goe: the tables %s and %s %w", b.table, pk2.table, ErrNotManyToMany)
	}
	b.sql.WriteString(mtmValue.table)
	b.sql.Write([]byte{32, 40})
	b.sql.WriteString(mtmValue.ids[pk1.table].attributeName)
	b.sql.WriteByte(44)
	b.sql.WriteString(mtmValue.ids[pk2.table].attributeName)
	b.sql.Write([]byte{41, 32})
	b.sql.WriteString("VALUES ")
	b.sql.WriteString("($1,$2)")
	b.argsAny = make([]any, v.Len())
	b.argsAny[0] = v.Index(0).Interface()
	b.argsAny[1] = v.Index(1).Interface()

	for i := 2; i < v.Len(); i++ {
		b.argsAny[i] = v.Index(i).Interface()
	}
	c := 1
	for i := 2; i <= v.Len()/2; i++ {
		b.sql.WriteString(fmt.Sprintf(",($%v,$%v)", i+c, i+c+1))
		c++
	}
	b.sql.WriteByte(59)
	return nil
}

func (b *builder) buildUpdate(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("UPDATE ")

	b.structColumns = make([]string, 0, len(b.args))
	b.attrNames = make([]string, 0, len(b.args))

	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
	b.sql.WriteString(" SET ")
	addrMap[b.args[0]].buildAttributeUpdate(b)

	a := b.args[1:]
	for i := range a {
		addrMap[a[i]].buildAttributeUpdate(b)
	}
}

func (b *builder) buildSet(value reflect.Value) {
	b.argsAny = make([]any, 0, len(b.attrNames))
	var c uint16 = 1
	buildSetField(value.FieldByName(b.structColumns[0]), b.attrNames[0], b, c)

	a := b.attrNames[1:]
	s := b.structColumns[1:]
	for i := range a {
		b.sql.WriteByte(44)
		c++
		buildSetField(value.FieldByName(s[i]), a[i], b, c)
	}
}

func buildSetField(valueField reflect.Value, fieldName string, b *builder, c uint16) {
	b.sql.WriteString(fmt.Sprintf("%v = $%v", fieldName, c))
	b.argsAny = append(b.argsAny, valueField.Interface())
	c++
}

func (b *builder) buildUpdateIn(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("UPDATE ")

	b.attrNames = make([]string, 0, len(b.args))
	b.tablesPk = make([]*pk, 2)

	b.table = addrMap[b.args[0]].getPrimaryKey().table
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
	b.tablesPk[1] = addrMap[b.args[1]].getPrimaryKey()
}

func (b *builder) buildSetIn() error {
	pk2 := b.tablesPk[1]

	mtm := pk2.fks[b.table]
	if mtm == nil {
		return fmt.Errorf("goe: the tables %s and %s %w", b.table, pk2.table, ErrNoMatchesTables)
	}

	if mtmValue, ok := mtm.(*manyToMany); ok {
		b.sql.WriteString(mtmValue.table)
		b.sql.WriteString(" SET ")
		b.sql.WriteString(fmt.Sprintf("%v = $1", mtmValue.ids[pk2.table].attributeName))
		return nil
	}
	return fmt.Errorf("goe: the tables %s and %s %w", b.table, pk2.table, ErrNotManyToMany)
}

func (b *builder) buildDelete(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("DELETE FROM ")
	b.sql.WriteString(addrMap[b.args[0]].getPrimaryKey().table)
}

func (b *builder) buildDeleteIn(addrMap map[uintptr]field) {
	//TODO: Set a drive type to share stm
	b.sql.WriteString("DELETE FROM ")

	b.tablesPk = make([]*pk, 2)

	b.table = addrMap[b.args[0]].getPrimaryKey().table
	b.tablesPk[0] = addrMap[b.args[0]].getPrimaryKey()
	b.tablesPk[1] = addrMap[b.args[1]].getPrimaryKey()
}

func (b *builder) buildSqlDeleteIn() (err error) {
	mtm := b.tablesPk[1].fks[b.table]
	if mtm == nil {
		return fmt.Errorf("goe: the tables %s and %s %w", b.table, b.tablesPk[1].table, ErrNoMatchesTables)
	}

	if mtmValue, ok := mtm.(*manyToMany); ok {
		b.sql.WriteString(mtmValue.table)
		err = b.buildWhereIn()
		b.sql.WriteByte(59)
		return err
	}
	return fmt.Errorf("goe: the tables %s and %s %w", b.table, b.tablesPk[1].table, ErrNotManyToMany)
}
