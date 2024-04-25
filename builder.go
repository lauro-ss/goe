package goe

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

const (
	querySELECT int8 = 1
	queryINSERT int8 = 2
	queryUPDATE int8 = 3
)

type builder struct {
	conn      conn
	sql       *strings.Builder
	args      []any
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

func (b *builder) buildSelect(addrMap map[string]any) Rows {
	//TODO: Set a drive type to share stm
	b.queue.add(&SELECT)

	//TODO Better Query
	for _, v := range b.args {
		addr := fmt.Sprintf("%p", v)
		switch atr := addrMap[addr].(type) {
		case *att:
			b.queue.add(createStatement(atr.name, ATT))
			b.tables.add(createStatement(atr.pk.table, TABLE))

			//TODO: Add a list pk?
			b.pks.add(atr.pk)
		case *pk:
			b.queue.add(createStatement(atr.name, ATT))
			b.tables.add(createStatement(atr.table, TABLE))

			//TODO: Add a list pk?
			b.pks.add(atr)
		}
	}

	b.queue.add(&FROM)
	return b
}

func (b *builder) Result(target any) {

	//generate query
	b.buildSql()

	// db.errors = nil
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}
	fmt.Println(b.sql)
	//b.handlerResult(value.Elem())
}

func (b *builder) buildSql() {
	switch b.queryType {
	case querySELECT:
		b.writeTables()
		buildSelect(b.sql, b.queue)
	case queryINSERT:
		break
	case queryUPDATE:
		break
	}
}

func (b *builder) writeTables() {
	b.queue.add(b.tables.get())
	if b.tables.size >= 1 {
		for table := b.tables.get(); table != nil; {
			writeJoins(table, b.pks, b.queue)
			table = b.tables.get()
		}
	}
}

func writeJoins(table *statement, pks *pkQueue, stQueue *statementQueue) {
	for pk := pks.get(); pk != nil; {
		switch fk := pk.fks[table.keyword].(type) {
		case *manyToOne:
			if fk.hasMany {
				stQueue.add(
					createStatement(fmt.Sprintf("inner join %v on (%v = %v)", table.keyword, pk.name, fk.id), JOIN),
				)

			}
		case *manyToMany:
			stQueue.add(
				createStatement(fmt.Sprintf("inner join %v on (%v = %v)", fk.table, pk.name, fk.ids[pk.table]), JOIN),
			)
			stQueue.add(
				createStatement(
					fmt.Sprintf(
						"inner join %v on (%v = %v)",
						table.keyword, fk.ids[table.keyword],
						pks.findPk(table.keyword).name), JOIN,
				),
			)

		}
		pk = pks.get()
	}
}

func (b *builder) handlerResult(value reflect.Value) {
	switch value.Kind() {
	case reflect.Slice:
		b.handlerQuery(value)
	case reflect.Struct:
		fmt.Println("struct")
	default:
		fmt.Println("default")
	}
}

func (b *builder) handlerQuery(value reflect.Value) {
	rows, err := b.conn.QueryContext(context.Background(), b.sql.String())

	//TODO: Better error
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	//Prepare dest for query
	c, err := rows.Columns()

	if err != nil {
		fmt.Println(err)
		return
	}

	dest := make([]any, len(c))
	for i := range dest {
		dest[i] = new(sql.RawBytes)
	}

	//Check the result target
	switch value.Type().Elem().Kind() {
	case reflect.Struct:
		err = mapStructQuery(rows, dest, value)

		//TODO: Better error
		if err != nil {
			fmt.Println(err)
			return
		}
	default:
		err = mapQuery(rows, dest, value)

		//TODO: Better error
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
