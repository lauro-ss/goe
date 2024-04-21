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
	queue     *queue
	tables    *queue
	queryType int8
}

func createBuilder(qt int8) *builder {
	return &builder{sql: &strings.Builder{}, queue: createQueue(), tables: createQueue(), queryType: qt}
}

func (b *builder) Result(target any) {

	//verifiy tables and joins

	//generate query
	b.buildSql()

	// db.errors = nil
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}

	b.handlerResult(value.Elem())
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
		//check joins
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
